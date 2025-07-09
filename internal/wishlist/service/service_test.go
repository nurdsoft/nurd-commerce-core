package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	cartclient "github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	productEntities "github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/repository"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockRepo := repository.NewMockRepository(ctrl)
	mockLogger := zap.NewExample().Sugar()
	mockConfig := cfg.Config{}
	mockProductClient := productclient.NewMockClient(ctrl)
	mockCartClient := cartclient.NewMockClient(ctrl)

	// Call the constructor
	svc := New(mockRepo, mockLogger, mockConfig, mockProductClient, mockCartClient)

	// Verify the service was created and is not nil
	assert.NotNil(t, svc, "Service should not be nil")

	// Optionally, type assertion to verify it's the correct type
	_, ok := svc.(*service)
	assert.True(t, ok, "Service should be of type *service")
}

func Test_service_AddToWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *productclient.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockProductClient := productclient.NewMockClient(ctrl)
		customerUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), customerUUID.String())
		svc := &service{
			repo:          mockRepo,
			log:           zap.NewExample().Sugar(),
			productClient: mockProductClient,
		}
		return svc, ctx, mockRepo, mockProductClient
	}

	t.Run("Valid request with a product ID", func(t *testing.T) {
		svc, ctx, mockRepo, mockProductClient := setup()
		productUUID := uuid.New()
		req := &entities.AddToWishlistRequest{
			Body: &entities.AddToWishlistRequestBody{
				Products: []entities.Product{
					{
						ProductID: productUUID,
					},
				},
			},
		}

		// Mock product client to return product details
		mockProductClient.EXPECT().GetProduct(ctx, &productEntities.GetProductRequest{
			ProductID: productUUID,
		}).Return(&productEntities.Product{
			ID: productUUID,
		}, nil).Times(1)

		// Mock repository to update wishlist with product IDs
		mockRepo.EXPECT().UpdateWishlist(ctx, meta.XCustomerID(ctx), []uuid.UUID{productUUID}).Return(nil).Times(1)

		err := svc.AddToWishlist(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Create product when it doesn't exist", func(t *testing.T) {
		svc, ctx, mockRepo, mockProductClient := setup()
		productUUID := uuid.New()
		productData := &entities.ProductData{
			Name: "Test Product",
		}

		req := &entities.AddToWishlistRequest{
			Body: &entities.AddToWishlistRequestBody{
				Products: []entities.Product{
					{
						ProductID:   productUUID,
						ProductData: productData,
					},
				},
			},
		}

		// Mock product client to return nil for GetProduct (product not found)
		mockProductClient.EXPECT().GetProduct(ctx, &productEntities.GetProductRequest{
			ProductID: productUUID,
		}).Return(nil, nil).Times(1)

		// Mock product client to create a product
		mockProductClient.EXPECT().CreateProduct(ctx, &productEntities.CreateProductRequest{
			Data: &productEntities.CreateProductRequestBody{
				ID:          &productUUID,
				Name:        productData.Name,
				Description: productData.Description,
				ImageURL:    productData.ImageURL,
				Attributes:  productData.Attributes,
			},
		}).Return(&productEntities.Product{
			ID: productUUID,
		}, nil).Times(1)

		// Mock repository to update wishlist
		mockRepo.EXPECT().UpdateWishlist(ctx, meta.XCustomerID(ctx), []uuid.UUID{productUUID}).Return(nil).Times(1)

		err := svc.AddToWishlist(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Product creation failure", func(t *testing.T) {
		svc, ctx, _, mockProductClient := setup()
		productUUID := uuid.New()
		productData := &entities.ProductData{
			Name: "Test Product",
		}

		req := &entities.AddToWishlistRequest{
			Body: &entities.AddToWishlistRequestBody{
				Products: []entities.Product{
					{
						ProductID:   productUUID,
						ProductData: productData,
					},
				},
			},
		}

		// Mock product client to return nil for GetProduct (product not found)
		mockProductClient.EXPECT().GetProduct(ctx, &productEntities.GetProductRequest{
			ProductID: productUUID,
		}).Return(nil, nil).Times(1)

		// Mock create product to return an error
		expectedErr := errors.New("failed to create product")
		mockProductClient.EXPECT().CreateProduct(ctx, &productEntities.CreateProductRequest{
			Data: &productEntities.CreateProductRequestBody{
				ID:          &productUUID,
				Name:        productData.Name,
				Description: productData.Description,
				ImageURL:    productData.ImageURL,
				Attributes:  productData.Attributes,
			},
		}).Return(nil, expectedErr).Times(1)

		// Repository should not be called since the product creation failed
		err := svc.AddToWishlist(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Product not found and no data to create new product", func(t *testing.T) {
		svc, ctx, _, mockProductClient := setup()
		productUUID := uuid.New()
		req := &entities.AddToWishlistRequest{
			Body: &entities.AddToWishlistRequestBody{
				Products: []entities.Product{
					{
						ProductID: productUUID,
						// Note: No ProductData provided here
					},
				},
			},
		}

		// Mock product client to return nil for GetProduct (product not found) without error
		mockProductClient.EXPECT().GetProduct(ctx, &productEntities.GetProductRequest{
			ProductID: productUUID,
		}).Return(nil, nil).Times(1)

		// No create product call expected since no product data was provided

		err := svc.AddToWishlist(ctx, req)

		// Should return a PRODUCT_NOT_FOUND API error
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		apiErr, _ := err.(*appErrors.APIError)
		assert.Equal(t, "Product not found.", apiErr.Error())
	})

	t.Run("No customer ID in context", func(t *testing.T) {
		svc, _, _, _ := setup()
		req := &entities.AddToWishlistRequest{
			Body: &entities.AddToWishlistRequestBody{
				Products: []entities.Product{
					{
						ProductID: uuid.New(),
					},
				},
			},
		}
		ctx := meta.WithXCustomerID(context.Background(), "")
		err := svc.AddToWishlist(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("Repository error is propagated", func(t *testing.T) {
		svc, ctx, mockRepo, mockProductClient := setup()
		productUUID := uuid.New()
		req := &entities.AddToWishlistRequest{
			Body: &entities.AddToWishlistRequestBody{
				Products: []entities.Product{
					{
						ProductID: productUUID,
					},
				},
			},
		}

		// Mock product client to return product details
		mockProductClient.EXPECT().GetProduct(ctx, &productEntities.GetProductRequest{
			ProductID: productUUID,
		}).Return(&productEntities.Product{
			ID: productUUID,
		}, nil).Times(1)

		// Mock repository to return an error
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().UpdateWishlist(ctx, meta.XCustomerID(ctx), []uuid.UUID{productUUID}).Return(expectedErr).Times(1)

		err := svc.AddToWishlist(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func Test_service_RemoveFromWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		customerUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), customerUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request with a product ID", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.RemoveFromWishlistRequest{
			ProductID: uuid.New(),
		}

		mockRepo.EXPECT().DeleteFromWishlist(ctx, meta.XCustomerID(ctx), req.ProductID).Return(nil).Times(1)
		err := svc.RemoveFromWishlist(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("No customer ID in context", func(t *testing.T) {
		svc, _, _ := setup()
		req := &entities.RemoveFromWishlistRequest{
			ProductID: uuid.New(),
		}
		ctx := meta.WithXCustomerID(context.Background(), "")
		err := svc.RemoveFromWishlist(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Customer ID is required.", err.Error())
	})

	t.Run("Repository error is propagated", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.RemoveFromWishlistRequest{
			ProductID: uuid.New(),
		}

		expectedErr := errors.New("database error")
		mockRepo.EXPECT().DeleteFromWishlist(ctx, meta.XCustomerID(ctx), req.ProductID).Return(expectedErr).Times(1)

		err := svc.RemoveFromWishlist(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func Test_service_GetWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		customerUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), customerUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.GetWishlistRequest{
			Limit: 10,
		}
		wishlist := []*entities.WishlistItem{
			{
				ProductID: uuid.New(),
			},
		}
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), req.Limit, "").Return(wishlist, "", int64(10), nil).Times(1)
		resp, err := svc.GetWishlist(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, wishlist, resp.Items)
		assert.Equal(t, int64(10), resp.Total)
	})

	t.Run("No customer ID in context", func(t *testing.T) {
		svc, _, _ := setup()
		req := &entities.GetWishlistRequest{}
		ctx := meta.WithXCustomerID(context.Background(), "")
		resp, err := svc.GetWishlist(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Nil(t, resp)
	})

	t.Run("Repository error is propagated", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.GetWishlistRequest{
			Limit:  10,
			Cursor: "",
		}

		expectedErr := errors.New("database error")
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), req.Limit, req.Cursor).Return(nil, "", int64(0), expectedErr).Times(1)

		resp, err := svc.GetWishlist(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("Returns next cursor when paginated results", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.GetWishlistRequest{
			Limit: 10,
		}
		wishlist := []*entities.WishlistItem{
			{
				ProductID: uuid.New(),
			},
		}
		expectedNextCursor := "next-page-token"

		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), req.Limit, req.Cursor).
			Return(wishlist, expectedNextCursor, int64(10), nil).Times(1)

		resp, err := svc.GetWishlist(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, wishlist, resp.Items)
		assert.Equal(t, expectedNextCursor, resp.NextCursor)
	})
}

func Test_service_BulkRemoveFromWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	customerUUID := uuid.New()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		ctx := meta.WithXCustomerID(context.Background(), customerUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("Successfully removes products from wishlist", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		productIDs := []uuid.UUID{uuid.New(), uuid.New()}
		req := &entities.BulkRemoveFromWishlistRequest{
			CustomerID: customerUUID,
			ProductIDs: productIDs,
		}

		// Note: The service passes req.CustomerID directly to the repository
		mockRepo.EXPECT().BulkRemoveFromWishlist(ctx, customerUUID, productIDs).Return(nil).Times(1)

		err := svc.BulkRemoveFromWishlist(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Handles repository errors", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		productIDs := []uuid.UUID{uuid.New(), uuid.New()}
		req := &entities.BulkRemoveFromWishlistRequest{
			CustomerID: customerUUID,
			ProductIDs: productIDs,
		}

		expectedErr := errors.New("repository error")
		// Note: The service passes req.CustomerID directly to the repository
		mockRepo.EXPECT().BulkRemoveFromWishlist(ctx, customerUUID, productIDs).Return(expectedErr).Times(1)

		err := svc.BulkRemoveFromWishlist(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func Test_service_GetMoreFromWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	customerUUID := uuid.New()

	setup := func() (*service, context.Context, *repository.MockRepository, *cartclient.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockCartClient := cartclient.NewMockClient(ctrl)
		ctx := meta.WithXCustomerID(context.Background(), customerUUID.String())
		svc := &service{
			repo:       mockRepo,
			log:        zap.NewExample().Sugar(),
			cartClient: mockCartClient,
		}
		return svc, ctx, mockRepo, mockCartClient
	}

	t.Run("Successfully gets more products from wishlist", func(t *testing.T) {
		svc, ctx, mockRepo, mockCartClient := setup()
		req := &entities.GetMoreFromWishlistRequest{}

		mockCartClient.EXPECT().GetCartItems(ctx).Return(&cartEntities.GetCartItemsResponse{
			Items: []cartEntities.CartItemDetail{
				{
					ProductID: uuid.New(),
					Quantity:  1,
				},
			},
		}, nil).Times(1)
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), 0, "").Return([]*entities.WishlistItem{
			{
				Id:         uuid.New(),
				CustomerID: customerUUID,
				ProductID:  uuid.New(),
			},
		}, "", int64(0), nil).Times(1)

		resp, err := svc.GetMoreFromWishlist(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("No customer ID in context", func(t *testing.T) {
		svc, _, _, _ := setup()
		req := &entities.GetMoreFromWishlistRequest{}
		ctx := meta.WithXCustomerID(context.Background(), "")

		resp, err := svc.GetMoreFromWishlist(ctx, req)

		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Nil(t, resp)
	})

	t.Run("Cart client returns error", func(t *testing.T) {
		svc, ctx, mockRepo, mockCartClient := setup()
		req := &entities.GetMoreFromWishlistRequest{}

		expectedErr := errors.New("cart service error")
		mockCartClient.EXPECT().GetCartItems(ctx).Return(nil, expectedErr).Times(1)
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), 0, "").Return([]*entities.WishlistItem{
			{
				Id:         uuid.New(),
				CustomerID: customerUUID,
				ProductID:  uuid.New(),
			},
		}, "", int64(0), nil).Times(1)

		resp, err := svc.GetMoreFromWishlist(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("Repository returns error", func(t *testing.T) {
		svc, ctx, mockRepo, mockCartClient := setup()
		req := &entities.GetMoreFromWishlistRequest{}

		mockCartClient.EXPECT().GetCartItems(ctx).Return(&cartEntities.GetCartItemsResponse{
			Items: []cartEntities.CartItemDetail{},
		}, nil).Times(1)

		expectedErr := errors.New("repository error")
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), 0, "").Return(nil, "", int64(0), expectedErr).Times(1)

		resp, err := svc.GetMoreFromWishlist(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("Empty wishlist returns nil", func(t *testing.T) {
		svc, ctx, mockRepo, mockCartClient := setup()
		req := &entities.GetMoreFromWishlistRequest{}

		mockCartClient.EXPECT().GetCartItems(ctx).Return(&cartEntities.GetCartItemsResponse{
			Items: []cartEntities.CartItemDetail{},
		}, nil).Times(1)

		// Return empty wishlist
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), 0, "").Return([]*entities.WishlistItem{}, "", int64(0), nil).Times(1)

		resp, err := svc.GetMoreFromWishlist(ctx, req)

		assert.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Filters out products already in cart", func(t *testing.T) {
		svc, ctx, mockRepo, mockCartClient := setup()
		req := &entities.GetMoreFromWishlistRequest{}

		// Create a product that will be in both cart and wishlist
		sharedProductID := uuid.New()

		// Another product only in wishlist
		wishlistOnlyProductID := uuid.New()

		mockCartClient.EXPECT().GetCartItems(ctx).Return(&cartEntities.GetCartItemsResponse{
			Items: []cartEntities.CartItemDetail{
				{
					ProductID: sharedProductID,
					Quantity:  1,
				},
			},
		}, nil).Times(1)

		// Create wishlist with two products - one in cart, one not in cart
		creationTime := time.Now()
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), 0, "").Return([]*entities.WishlistItem{
			{
				Id:         uuid.New(),
				CustomerID: customerUUID,
				ProductID:  sharedProductID,
				CreatedAt:  creationTime,
			},
			{
				Id:         uuid.New(),
				CustomerID: customerUUID,
				ProductID:  wishlistOnlyProductID,
				CreatedAt:  creationTime,
			},
		}, "", int64(0), nil).Times(1)

		resp, err := svc.GetMoreFromWishlist(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// Only the product not in cart should be returned
		assert.Equal(t, 1, len(resp.Items))
		assert.Equal(t, wishlistOnlyProductID, resp.Items[0].ProductID)
	})

	t.Run("Sorts items by created_at in descending order", func(t *testing.T) {
		svc, ctx, mockRepo, mockCartClient := setup()
		req := &entities.GetMoreFromWishlistRequest{}

		// Create timestamps in ascending order
		now := time.Now()
		olderTime := now.Add(-2 * time.Hour)
		newerTime := now.Add(-1 * time.Hour)

		// Two product IDs
		product1ID := uuid.New()
		product2ID := uuid.New()

		mockCartClient.EXPECT().GetCartItems(ctx).Return(&cartEntities.GetCartItemsResponse{
			Items: []cartEntities.CartItemDetail{},
		}, nil).Times(1)

		// Return wishlist items with different timestamps (older first, newer second)
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), 0, "").Return([]*entities.WishlistItem{
			{
				Id:         uuid.New(),
				CustomerID: customerUUID,
				ProductID:  product1ID,
				CreatedAt:  olderTime, // Older item
			},
			{
				Id:         uuid.New(),
				CustomerID: customerUUID,
				ProductID:  product2ID,
				CreatedAt:  newerTime, // Newer item
			},
		}, "", int64(0), nil).Times(1)

		resp, err := svc.GetMoreFromWishlist(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 2, len(resp.Items))

		// Verify that items are sorted by CreatedAt in descending order (newer first)
		assert.Equal(t, product2ID, resp.Items[0].ProductID) // Newer item should be first
		assert.Equal(t, product1ID, resp.Items[1].ProductID) // Older item should be second

		// Explicitly verify the timestamps are in the correct order
		assert.True(t, resp.Items[0].CreatedAt.After(resp.Items[1].CreatedAt))
	})

	t.Run("Empty cart results in all wishlist items", func(t *testing.T) {
		svc, ctx, mockRepo, mockCartClient := setup()
		req := &entities.GetMoreFromWishlistRequest{}

		// Create a product that will be in both cart and wishlist
		sharedProductID := uuid.New()

		// Another product only in wishlist
		wishlistOnlyProductID := uuid.New()

		mockCartClient.EXPECT().GetCartItems(ctx).Return(&cartEntities.GetCartItemsResponse{
			Items: []cartEntities.CartItemDetail{},
		}, nil).Times(1)

		// Create wishlist with two products - one in cart, one not in cart
		creationTime := time.Now()
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), 0, "").Return([]*entities.WishlistItem{
			{
				Id:         uuid.New(),
				CustomerID: customerUUID,
				ProductID:  sharedProductID,
				CreatedAt:  creationTime,
			},
			{
				Id:         uuid.New(),
				CustomerID: customerUUID,
				ProductID:  wishlistOnlyProductID,
				CreatedAt:  creationTime,
			},
		}, "", int64(0), nil).Times(1)

		resp, err := svc.GetMoreFromWishlist(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// Both products should be returned since cart is empty
		assert.Equal(t, 2, len(resp.Items))
		assert.Equal(t, sharedProductID, resp.Items[0].ProductID)
		assert.Equal(t, wishlistOnlyProductID, resp.Items[1].ProductID)
	})
}

func Test_service_GetWishlistProductTimestamps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepository(ctrl)
	svc := &service{
		repo: mockRepo,
		log:  zap.NewExample().Sugar(),
	}
	customerUUID := uuid.New()
	ctx := meta.WithXCustomerID(context.Background(), customerUUID.String())
	productIDs := []uuid.UUID{uuid.New(), uuid.New()}

	req := &entities.GetWishlistProductTimestampsRequest{
		Body: &entities.GetWishlistProductTimestampsRequestBody{
			ProductIDs: productIDs,
		},
	}

	t.Run("Returns timestamps successfully", func(t *testing.T) {

		expected := map[string]time.Time{
			productIDs[0].String(): time.Now().Add(-time.Hour),
			productIDs[1].String(): time.Now(),
		}

		expectedResp := &entities.GetWishlistProductTimestampsResponse{
			Timestamps: expected,
		}

		mockRepo.EXPECT().
			GetWishlistProductTimestamps(customerUUID.String(), productIDs).
			Return(expected, nil).Times(1)

		resp, err := svc.GetWishlistProductTimestamps(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("Returns error from repository", func(t *testing.T) {
		expectedErr := errors.New("repository failure")

		mockRepo.EXPECT().
			GetWishlistProductTimestamps(customerUUID.String(), productIDs).
			Return(nil, expectedErr).Times(1)

		resp, err := svc.GetWishlistProductTimestamps(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Returns error when customer ID missing", func(t *testing.T) {
		ctx := meta.WithXCustomerID(context.Background(), "")
		resp, err := svc.GetWishlistProductTimestamps(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}
