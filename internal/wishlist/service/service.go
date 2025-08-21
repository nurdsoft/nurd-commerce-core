package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	productEntities "github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/wishlist/errors"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/repository"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"go.uber.org/zap"
)

type Service interface {
	AddToWishlist(ctx context.Context, req *entities.AddToWishlistRequest) error
	RemoveFromWishlist(ctx context.Context, req *entities.RemoveFromWishlistRequest) error
	GetWishlist(ctx context.Context, req *entities.GetWishlistRequest) (*entities.GetWishlistResponse, error)
	BulkRemoveFromWishlist(ctx context.Context, req *entities.BulkRemoveFromWishlistRequest) error
	GetMoreFromWishlist(ctx context.Context, req *entities.GetMoreFromWishlistRequest) (*entities.GetWishlistResponse, error)
	GetWishlistProductTimestamps(ctx context.Context, req *entities.GetWishlistProductTimestampsRequest) (*entities.GetWishlistProductTimestampsResponse, error)
}

type service struct {
	repo          repository.Repository
	log           *zap.SugaredLogger
	config        cfg.Config
	productClient productclient.Client
	cartClient    cartclient.Client
}

func New(repo repository.Repository, logger *zap.SugaredLogger, config cfg.Config,
	productClient productclient.Client, cartClient cartclient.Client) Service {
	return &service{
		repo:          repo,
		log:           logger,
		config:        config,
		productClient: productClient,
		cartClient:    cartClient,
	}
}

// swagger:route PUT /wishlist wishlist AddToWishlistRequest
//
// # Add Products to Wishlist
// ### Add products to the customer's wishlist
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefaultResponse Product added to wishlist successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) AddToWishlist(ctx context.Context, req *entities.AddToWishlistRequest) error {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	var productIDs []uuid.UUID
	for _, productReq := range req.Body.Products {
		product, err := s.productClient.GetProduct(ctx, &productEntities.GetProductRequest{
			ProductID: productReq.ProductID,
		})

		if (err != nil || product == nil) && productReq.ProductData != nil {
			productData := *productReq.ProductData
			product, err = s.productClient.CreateProduct(ctx, &productEntities.CreateProductRequest{
				Data: &productEntities.CreateProductRequestBody{
					ID:          &productReq.ProductID,
					Name:        productData.Name,
					Description: productData.Description,
					ImageURL:    productData.ImageURL,
					Attributes:  productData.Attributes,
				},
			})
			if err != nil {
				return err
			}
		}
		if product == nil {
			return moduleErrors.NewAPIError("PRODUCT_NOT_FOUND")
		}
		productIDs = append(productIDs, product.ID)
	}

	err := s.repo.UpdateWishlist(ctx, customerID, productIDs)
	if err != nil {
		return err
	}

	return nil
}

// swagger:route DELETE /wishlist/{product_id} wishlist RemoveFromWishlistRequest
//
// # Remove Product from Wishlist
// ### Remove a product from the customer's wishlist
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefaultResponse Product removed from wishlist successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) RemoveFromWishlist(ctx context.Context, req *entities.RemoveFromWishlistRequest) error {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	err := s.repo.DeleteFromWishlist(ctx, customerID, req.ProductID)
	if err != nil {
		return err
	}

	return nil
}

// swagger:route GET /wishlist wishlist GetWishlistRequest
//
// # Get Wishlist
// ### Get the products in customer's wishlist
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetWishlistResponse Wishlist fetched successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) GetWishlist(ctx context.Context, req *entities.GetWishlistRequest) (*entities.GetWishlistResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}
	var items []*entities.WishlistItem

	items, nextCursor, total, err := s.repo.GetWishlist(ctx, customerID, req.Limit, req.Cursor)
	if err != nil {
		return nil, err
	}

	return &entities.GetWishlistResponse{
		Items:      items,
		NextCursor: nextCursor,
		Total:      total,
	}, nil
}

func (s *service) BulkRemoveFromWishlist(ctx context.Context, req *entities.BulkRemoveFromWishlistRequest) error {
	err := s.repo.BulkRemoveFromWishlist(ctx, req.CustomerID, req.ProductIDs)
	if err != nil {
		return err
	}

	return nil
}

// swagger:route GET /wishlist/more wishlist GetMoreFromWishlistRequest
//
// # Get More from Wishlist
// ### Get products from a customer's wishlist that aren't already in the cart
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetWishlistResponse Wishlist fetched successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) GetMoreFromWishlist(ctx context.Context, req *entities.GetMoreFromWishlistRequest) (*entities.GetWishlistResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)
	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	cartItems, _ := s.cartClient.GetCartItems(ctx)

	var cartProductIDs []uuid.UUID
	if cartItems != nil {
		for _, item := range cartItems.Items {
			cartProductIDs = append(cartProductIDs, item.ProductID)
		}
	}

	wishlistItems, nextCursor, total, err := s.repo.GetMoreFromWishlist(ctx, customerID, req.Limit, req.Cursor, cartProductIDs)
	if err != nil {
		return nil, err
	}

	if len(wishlistItems) == 0 {
		return nil, nil
	}

	return &entities.GetWishlistResponse{
		Items:      wishlistItems,
		NextCursor: nextCursor,
		Total:      total,
	}, nil

}

// swagger:route POST /wishlist/timestamps wishlist GetWishlistProductTimestampsRequest
//
// # Get Wishlist Product Timestamps
// ### Retrieve the timestamps when products were added to the customer's wishlist
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetWishlistProductTimestampsResponse Timestamps retrieved successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetWishlistProductTimestamps(ctx context.Context, req *entities.GetWishlistProductTimestampsRequest) (*entities.GetWishlistProductTimestampsResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	timestamps, err := s.repo.GetWishlistProductTimestamps(customerID, req.Body.ProductIDs)
	if err != nil {
		return nil, err
	}

	return &entities.GetWishlistProductTimestampsResponse{
		Timestamps: timestamps,
	}, nil
}
