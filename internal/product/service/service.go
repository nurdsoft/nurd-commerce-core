package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/repository"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"go.uber.org/zap"
)

type Service interface {
	CreateProduct(ctx context.Context, req *entities.CreateProductRequest) (*entities.Product, error)
	GetProduct(ctx context.Context, req *entities.GetProductRequest) (*entities.Product, error)
	GetProductsByIDs(ctx context.Context, ids []string) ([]entities.Product, error)
	UpdateProduct(ctx context.Context, req *entities.UpdateProductRequest) error
	CreateProductVariant(ctx context.Context, req *entities.CreateProductVariantRequest) (*entities.ProductVariant, error)
	GetProductVariant(ctx context.Context, req *entities.GetProductVariantRequest) (*entities.ProductVariant, error)
	GetProductVariantByID(ctx context.Context, variantID string) (*entities.ProductVariant, error)
	ListProductVariants(ctx context.Context, req *entities.ListProductVariantsRequest) (*entities.ListProductVariantsResponse, error)
}

type service struct {
	repo   repository.Repository
	log    *zap.SugaredLogger
	config cfg.Config
}

func New(
	repo repository.Repository,
	logger *zap.SugaredLogger,
	config cfg.Config,
) Service {
	return &service{
		repo:   repo,
		log:    logger,
		config: config,
	}
}

// swagger:route POST /product products CreateProductRequest
//
// # Create Product
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetProductResponse Product created successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) CreateProduct(ctx context.Context, req *entities.CreateProductRequest) (*entities.Product, error) {

	if req.Data.ID == nil {
		id := uuid.New()
		req.Data.ID = &id
	}

	product := &entities.Product{
		ID:          *req.Data.ID,
		Name:        req.Data.Name,
		Description: req.Data.Description,
		ImageURL:    req.Data.ImageURL,
		Attributes:  req.Data.Attributes,
	}

	createdProduct, err := s.repo.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	return createdProduct, nil
}

// swagger:route GET /product/{product_id} products GetProductRequest
//
// # Get product details
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetProductResponse
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) GetProduct(ctx context.Context, req *entities.GetProductRequest) (*entities.Product, error) {
	product, err := s.repo.FindByID(ctx, req.ProductID.String())
	if err != nil {
		return nil, err
	}
	return product, nil
}

// swagger:route POST /product/{product_id}/variant products CreateProductVariantRequest
//
// # Create Product Variant
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetProductVariantResponse
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) CreateProductVariant(ctx context.Context, req *entities.CreateProductVariantRequest) (*entities.ProductVariant, error) {
	product, err := s.repo.FindByID(ctx, req.ProductID.String())
	if err != nil {
		return nil, err
	}
	existingVariant, err := s.repo.FindVariantBySKU(ctx, req.Data.SKU)
	if err != nil || existingVariant == nil {
		newVariant := &entities.ProductVariant{
			ID:            uuid.New(),
			ProductID:     product.ID,
			SKU:           req.Data.SKU,
			Name:          req.Data.Name,
			Description:   req.Data.Description,
			ImageURL:      req.Data.ImageURL,
			Price:         req.Data.Price,
			Currency:      req.Data.Currency,
			Length:        req.Data.Length,
			Height:        req.Data.Height,
			Weight:        req.Data.Weight,
			Width:         req.Data.Width,
			Attributes:    req.Data.Attributes,
			StripeTaxCode: req.Data.StripeTaxCode,
		}
		createdVariant, err := s.repo.CreateVariant(ctx, newVariant)
		if err != nil {
			return nil, err
		}
		return createdVariant, nil
	} else {
		err := s.repo.UpdateVariant(ctx, map[string]interface{}{
			"name":        req.Data.Name,
			"description": req.Data.Description,
			"image_url":   req.Data.ImageURL,
			"price":       req.Data.Price,
			"currency":    req.Data.Currency,
			"length":      req.Data.Length,
			"height":      req.Data.Height,
			"weight":      req.Data.Weight,
			"attributes":  req.Data.Attributes,
		}, existingVariant.ID.String())
		if err != nil {
			return nil, err
		}
		updatedVariant, err := s.repo.FindVariantBySKU(ctx, req.Data.SKU)
		return updatedVariant, nil
	}
}

// swagger:route GET /product/variant/{sku} products GetProductVariantRequest
//
// # Get Product Variant
// ### Get product variant details by SKU
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetProductVariantResponse
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) GetProductVariant(ctx context.Context, req *entities.GetProductVariantRequest) (*entities.ProductVariant, error) {
	productVariant, err := s.repo.FindVariantBySKU(ctx, req.SKU)
	if err != nil {
		return nil, err
	}
	return productVariant, nil
}

func (s *service) GetProductVariantByID(ctx context.Context, variantID string) (*entities.ProductVariant, error) {
	productVariant, err := s.repo.FindVariantByID(ctx, variantID)
	if err != nil {
		return nil, err
	}
	return productVariant, nil
}

func (s *service) UpdateProduct(ctx context.Context, req *entities.UpdateProductRequest) error {
	err := s.repo.Update(ctx, map[string]interface{}{
		"salesforce_id":                 req.Data.SalesforceID,
		"salesforce_pricebook_entry_id": req.Data.SalesforcePricebookEntryId,
	}, req.ProductID.String())
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetProductsByIDs(ctx context.Context, ids []string) ([]entities.Product, error) {
	products, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return products, nil
}

// swagger:route GET /product/variants products ListProductVariantsRequest
//
// # List Product Variants
// ### Get a paginated list of product variants with optional filtering
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: ListProductVariantsResponse
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) ListProductVariants(ctx context.Context, req *entities.ListProductVariantsRequest) (*entities.ListProductVariantsResponse, error) {
	response, err := s.repo.ListVariants(ctx, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
