package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/repository"
	userEntities "github.com/nurdsoft/nurd-commerce-core/internal/user/entities"
	userClient "github.com/nurdsoft/nurd-commerce-core/internal/user/userclient"
	salesforce "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/client"
	sfEntities "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/entities"
	shipengineClient "github.com/nurdsoft/nurd-commerce-core/internal/vendors/shipengine/client"
	shipengineEntities "github.com/nurdsoft/nurd-commerce-core/internal/vendors/shipengine/entities"
	stripeClient "github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/client"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/nurdsoft/nurd-commerce-core/shared/nullable"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_service_UpdateCartItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (
		*service, context.Context,
		*repository.MockRepository,
		*repository.MockTransaction,
		*userClient.MockClient,
		*salesforce.MockClient,
		*cache.MockCache,
	) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockTx := repository.NewMockTransaction(ctrl)
		userClient := userClient.NewMockClient(ctrl)
		mockSfClient := salesforce.NewMockClient(ctrl)
		userUUID := uuid.New()
		mockCache := cache.NewMockCache(ctrl)
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:             mockRepo,
			log:              zap.NewExample().Sugar(),
			userClient:       userClient,
			salesforceClient: mockSfClient,
			cache:            mockCache,
		}
		return svc, ctx, mockRepo, mockTx, userClient, mockSfClient, mockCache
	}

	t.Run("Valid request with new item", func(t *testing.T) {
		svc, ctx, mockRepo, mockTx, mockUserClient, mockSfClient, mockCache := setup()
		req := &entities.UpdateCartItemRequest{
			Item: &entities.UpdateCartItemRequestBody{
				BookId:   uuid.New().String(),
				Format:   entities.BookFormatPaperback,
				Quantity: 1,
				Price:    decimal.NewFromFloat(12.34),
			},
		}
		mockCart := &entities.Cart{Id: uuid.New(), Status: entities.Active, UpdatedAt: time.Now()}

		mockRepo.EXPECT().BeginTransaction(gomock.Any()).Return(mockTx, nil)
		mockRepo.EXPECT().GetActiveCart(gomock.Any(), gomock.Any()).Return(mockCart, nil)
		mockRepo.EXPECT().GetCartItem(gomock.Any(), gomock.Any(), req.Item.BookId, req.Item.Format).Return(nil, nil)
		mockRepo.EXPECT().AddCartItem(gomock.Any(), gomock.Any(), gomock.Any(), req.Item).Return(&entities.CartItem{}, nil)
		mockTx.EXPECT().Commit().Return(&gorm.DB{Error: nil})
		mockUserClient.EXPECT().GetAddresses(gomock.Any()).Return(&userEntities.GetAllAddressResponse{}, nil).AnyTimes()
		mockRepo.EXPECT().GetSalesforceProductsByBookID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockSfClient.EXPECT().CreateProduct(gomock.Any(), gomock.Any()).Return(&sfEntities.CreateSFProductResponse{
			ID: "123",
		}, nil).AnyTimes()
		mockSfClient.EXPECT().CreatePriceBookEntry(gomock.Any(), gomock.Any()).Return(&sfEntities.CreateSFPriceBookEntryResponse{
			ID: "123",
		}, nil).AnyTimes()
		mockRepo.EXPECT().CreateSalesforceProduct(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockCache.EXPECT().DeleteByPattern(context.Background(), gomock.Any()).Return(nil).AnyTimes()

		_, err := svc.UpdateCartItem(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Error during transaction start", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _, _ := setup()
		req := &entities.UpdateCartItemRequest{
			Item: &entities.UpdateCartItemRequestBody{
				BookId:   uuid.New().String(),
				Format:   entities.BookFormatPaperback,
				Quantity: 1,
				Price:    decimal.NewFromFloat(12.34),
			},
		}

		mockRepo.EXPECT().BeginTransaction(ctx).Return(nil, errors.New("transaction error"))

		_, err := svc.UpdateCartItem(ctx, req)
		assert.Error(t, err)
	})

	t.Run("Valid request with existing item", func(t *testing.T) {
		svc, ctx, mockRepo, mockTx, _, _, mockCache := setup()
		req := &entities.UpdateCartItemRequest{
			Item: &entities.UpdateCartItemRequestBody{
				BookId:   uuid.New().String(),
				Format:   entities.BookFormatPaperback,
				Quantity: 2,
				Price:    decimal.NewFromFloat(12.34),
			},
		}
		mockCart := &entities.Cart{Id: uuid.New(), Status: entities.Active, UpdatedAt: time.Now()}
		mockItem := &entities.CartItem{Id: uuid.New(), Quantity: 1, Price: decimal.NewFromFloat(10.00)}

		mockRepo.EXPECT().BeginTransaction(ctx).Return(mockTx, nil)
		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(mockCart, nil)
		mockRepo.EXPECT().GetCartItem(ctx, gomock.Any(), req.Item.BookId, req.Item.Format).Return(mockItem, nil)
		mockRepo.EXPECT().UpdateCartItem(ctx, gomock.Any(), gomock.Any(), req.Item.Quantity, req.Item.Price).Return(nil)
		mockTx.EXPECT().Commit().Return(&gorm.DB{Error: nil})
		mockCache.EXPECT().DeleteByPattern(context.Background(), gomock.Any()).Return(nil).AnyTimes()

		_, err := svc.UpdateCartItem(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Valid request with new item and zero quantity", func(t *testing.T) {
		svc, ctx, mockRepo, mockTx, _, _, mockCache := setup()
		req := &entities.UpdateCartItemRequest{
			Item: &entities.UpdateCartItemRequestBody{
				BookId:   uuid.New().String(),
				Format:   entities.BookFormatPaperback,
				Quantity: 0,
				Price:    decimal.NewFromFloat(12.34),
			},
		}
		mockCart := &entities.Cart{Id: uuid.New(), Status: entities.Active, UpdatedAt: time.Now()}

		mockRepo.EXPECT().BeginTransaction(gomock.Any()).Return(mockTx, nil)
		mockRepo.EXPECT().GetActiveCart(gomock.Any(), gomock.Any()).Return(mockCart, nil)
		mockRepo.EXPECT().GetCartItem(gomock.Any(), gomock.Any(), req.Item.BookId, req.Item.Format).Return(nil, nil)
		mockTx.EXPECT().Commit().Return(&gorm.DB{Error: nil})
		mockCache.EXPECT().DeleteByPattern(context.Background(), gomock.Any()).Return(nil).AnyTimes()

		_, err := svc.UpdateCartItem(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Valid request with existing item and zero quantity", func(t *testing.T) {
		svc, ctx, mockRepo, mockTx, _, _, mockCache := setup()
		req := &entities.UpdateCartItemRequest{
			Item: &entities.UpdateCartItemRequestBody{
				BookId:   uuid.New().String(),
				Format:   entities.BookFormatPaperback,
				Quantity: 0,
				Price:    decimal.NewFromFloat(12.34),
			},
		}
		mockCart := &entities.Cart{Id: uuid.New(), Status: entities.Active, UpdatedAt: time.Now()}
		mockItem := &entities.CartItem{Id: uuid.New(), Quantity: 1, Price: decimal.NewFromFloat(10.00)}

		mockRepo.EXPECT().BeginTransaction(ctx).Return(mockTx, nil)
		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(mockCart, nil)
		mockRepo.EXPECT().GetCartItem(ctx, gomock.Any(), req.Item.BookId, req.Item.Format).Return(mockItem, nil)
		mockRepo.EXPECT().RemoveCartItem(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockTx.EXPECT().Commit().Return(&gorm.DB{Error: nil})
		mockCache.EXPECT().DeleteByPattern(context.Background(), gomock.Any()).Return(nil).AnyTimes()

		_, err := svc.UpdateCartItem(ctx, req)
		assert.NoError(t, err)
	})
}

func Test_service_GetCartItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("No active cart found", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(nil, nil)

		resp, err := svc.GetCartItems(ctx)
		assert.NoError(t, err)
		assert.Empty(t, resp.Items)
	})

	t.Run("Error retrieving active cart", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(nil, errors.New("database error"))

		_, err := svc.GetCartItems(ctx)
		assert.Error(t, err)
	})
}

func Test_service_RemoveCartItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *cache.MockCache) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		mockCache := cache.NewMockCache(ctrl)
		svc := &service{
			repo:  mockRepo,
			log:   zap.NewExample().Sugar(),
			cache: mockCache,
		}
		return svc, ctx, mockRepo, mockCache
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo, mockCache := setup()
		itemID := uuid.New().String()
		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(&entities.Cart{Id: uuid.New(), Status: entities.Active}, nil)
		mockRepo.EXPECT().RemoveCartItem(ctx, gomock.Any(), itemID).Return(nil)
		mockCache.EXPECT().DeleteByPattern(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		err := svc.RemoveCartItem(ctx, itemID)
		assert.NoError(t, err)
	})

	t.Run("Error removing cart item", func(t *testing.T) {
		svc, ctx, mockRepo, _ := setup()
		itemID := uuid.New().String()
		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(&entities.Cart{Id: uuid.New(), Status: entities.Active}, nil)
		mockRepo.EXPECT().RemoveCartItem(ctx, gomock.Any(), itemID).Return(errors.New("database error"))

		err := svc.RemoveCartItem(ctx, itemID)
		assert.Error(t, err)
	})
}

func Test_service_ClearCartItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(&entities.Cart{Id: uuid.New(), Status: entities.Active}, nil)
		mockRepo.EXPECT().UpdateCartStatus(ctx, gomock.Any(), gomock.Any(), "cleared").Return(nil)

		err := svc.ClearCartItems(ctx)
		assert.NoError(t, err)
	})

	t.Run("Error clearing cart items", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(&entities.Cart{Id: uuid.New(), Status: entities.Active}, nil)
		mockRepo.EXPECT().UpdateCartStatus(ctx, gomock.Any(), gomock.Any(), "cleared").Return(errors.New("database error"))

		err := svc.ClearCartItems(ctx)
		assert.Error(t, err)
	})
}

func Test_service_GetShippingRate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *shipengineClient.MockClient, *userClient.MockClient, *cache.MockCache) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockVendorClient := shipengineClient.NewMockClient(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockCache := cache.NewMockCache(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:             mockRepo,
			shipengineClient: mockVendorClient,
			log:              zap.NewExample().Sugar(),
			userClient:       mockUserClient,
			cache:            mockCache,
		}
		return svc, ctx, mockRepo, mockVendorClient, mockUserClient, mockCache
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo, mockVendorClient, mockUserClient, mockCache := setup()
		req := &entities.GetShippingRateRequest{
			Body: &entities.GetShippingRateRequestBody{
				AddressUUID: uuid.New(),
			},
		}

		mockUserClient.EXPECT().GetAddressByUUID(ctx, &userEntities.GetAddressRequest{
			AddressUUID: req.Body.AddressUUID,
		}).Return(
			&userEntities.GetAddressResponse{
				Address: userEntities.UserAddress{
					City:    "City",
					State:   "State",
					Country: "US",
					ZipCode: "12345",
				},
			}, nil,
		)

		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(&entities.Cart{
			Id:        uuid.New(),
			Status:    entities.Active,
			UpdatedAt: time.Now().Add(-time.Hour),
		}, nil)
		mockRepo.EXPECT().GetCartItems(ctx, gomock.Any()).Return([]entities.CartItem{
			{
				Id:       uuid.New(),
				Format:   entities.BookFormatPaperback,
				Price:    decimal.NewFromFloat(10.00),
				Quantity: 1,
				Length:   decimal.NewFromFloat(8.0),
				Width:    decimal.NewFromFloat(5.0),
				Height:   decimal.NewFromFloat(10.0),
				Weight:   decimal.NewFromFloat(2.0),
			},
		}, nil)

		mockCache.EXPECT().Get(ctx, gomock.Any()).Return(nil, nil).AnyTimes()

		mockVendorClient.EXPECT().GetRatesEstimate(ctx,
			shipengineEntities.ShippingAddress{
				City:    "La Vergne",
				State:   "TN",
				Zip:     "37086",
				Country: "US",
			},
			shipengineEntities.ShippingAddress{
				City:    "City",
				State:   "State",
				Zip:     "12345",
				Country: "US",
			},
			shipengineEntities.Dimensions{
				Height: decimal.NewFromFloat(10.0),
				Width:  decimal.NewFromFloat(5.0),
				Length: decimal.NewFromFloat(8.0),
				Weight: decimal.NewFromFloat(2.0),
			},
		).Return([]shipengineEntities.EstimateRatesResponse{
			{
				ShippingAmount: shipengineEntities.ShippingAmount{
					Amount:   100,
					Currency: "USD",
				},
				CarrierFriendlyName:   "Carrier",
				CarrierCode:           "CarrierCode",
				ServiceType:           "ServiceType",
				EstimatedDeliveryDate: nullable.NewNullTime(time.Now()),
			},
		}, nil)

		mockRepo.EXPECT().CreateCartShippingRates(ctx, gomock.Any()).Return(nil)
		mockCache.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.GetShippingRate(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Rates, 1)
		assert.Equal(t, "Carrier", resp.Rates[0].CarrierName)
	})

	t.Run("Error retrieving user address", func(t *testing.T) {
		svc, ctx, _, _, mockUserClient, _ := setup()
		req := &entities.GetShippingRateRequest{
			Body: &entities.GetShippingRateRequestBody{
				AddressUUID: uuid.New(),
			},
		}

		mockUserClient.EXPECT().GetAddressByUUID(ctx, &userEntities.GetAddressRequest{
			AddressUUID: req.Body.AddressUUID,
		}).Return(nil, errors.New("database error"))

		_, err := svc.GetShippingRate(ctx, req)
		assert.Error(t, err)
	})

	t.Run("Error retrieving shipping rates", func(t *testing.T) {
		svc, ctx, mockRepo, mockVendorClient, mockUserClient, mockCache := setup()
		req := &entities.GetShippingRateRequest{
			Body: &entities.GetShippingRateRequestBody{
				AddressUUID: uuid.New(),
			},
		}

		mockUserClient.EXPECT().GetAddressByUUID(ctx, &userEntities.GetAddressRequest{
			AddressUUID: req.Body.AddressUUID,
		}).Return(
			&userEntities.GetAddressResponse{
				Address: userEntities.UserAddress{
					City:    "City",
					State:   "State",
					Country: "US",
					ZipCode: "12345",
				},
			}, nil,
		)

		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(&entities.Cart{
			Id:        uuid.New(),
			Status:    entities.Active,
			UpdatedAt: time.Now().Add(-time.Hour),
		}, nil)

		mockRepo.EXPECT().GetCartItems(ctx, gomock.Any()).Return([]entities.CartItem{
			{
				Id:       uuid.New(),
				Format:   entities.BookFormatPaperback,
				Price:    decimal.NewFromFloat(10.00),
				Quantity: 1,
				Length:   decimal.NewFromFloat(8.0),
				Width:    decimal.NewFromFloat(5.0),
				Height:   decimal.NewFromFloat(10.0),
				Weight:   decimal.NewFromFloat(2.0),
			},
		}, nil)

		mockCache.EXPECT().Get(ctx, gomock.Any()).Return(nil, nil).AnyTimes()

		mockVendorClient.EXPECT().GetRatesEstimate(ctx,
			shipengineEntities.ShippingAddress{
				City:    "La Vergne",
				State:   "TN",
				Zip:     "37086",
				Country: "US",
			},
			shipengineEntities.ShippingAddress{
				City:    "City",
				State:   "State",
				Zip:     "12345",
				Country: "US",
			},
			shipengineEntities.Dimensions{
				Height: decimal.NewFromFloat(10.0),
				Width:  decimal.NewFromFloat(5.0),
				Length: decimal.NewFromFloat(8.0),
				Weight: decimal.NewFromFloat(2.0),
			},
		).Return(nil, errors.New("shipping rate error"))

		_, err := svc.GetShippingRate(ctx, req)
		assert.Error(t, err)
	})
}

func Test_service_GetTaxRate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *stripeClient.MockClient, *userClient.MockClient, *cache.MockCache) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		mockVendorClient := stripeClient.NewMockClient(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockCache := cache.NewMockCache(ctrl)
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:         mockRepo,
			stripeClient: mockVendorClient,
			log:          zap.NewExample().Sugar(),
			userClient:   mockUserClient,
			cache:        mockCache,
		}
		return svc, ctx, mockRepo, mockVendorClient, mockUserClient, mockCache
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo, mockVendorClient, mockUserClient, mockCache := setup()
		req := &entities.GetTaxRateRequest{
			Body: &entities.GetTaxRateRequestBody{
				AddressUUID: uuid.New(),
			},
		}

		mockUserClient.EXPECT().GetAddressByUUID(ctx, &userEntities.GetAddressRequest{
			AddressUUID: req.Body.AddressUUID,
		}).Return(
			&userEntities.GetAddressResponse{
				Address: userEntities.UserAddress{
					City:    "City",
					State:   "State",
					Country: "US",
					ZipCode: "37086",
				},
			}, nil,
		)

		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(&entities.Cart{
			Id:        uuid.New(),
			Status:    entities.Active,
			UpdatedAt: time.Now().Add(-time.Hour),
		}, nil)

		mockRepo.EXPECT().GetCartItems(ctx, gomock.Any()).Return([]entities.CartItem{
			{
				Id:       uuid.New(),
				Format:   entities.BookFormatPaperback,
				Price:    decimal.NewFromFloat(10.00),
				Quantity: 1,
				Length:   decimal.NewFromFloat(8.0),
				Width:    decimal.NewFromFloat(5.0),
				Height:   decimal.NewFromFloat(10.0),
				Weight:   decimal.NewFromFloat(2.0),
				Isbn:     "123456",
			},
		}, nil)

		mockRepo.EXPECT().GetShippingRate(ctx, gomock.Any()).Return(&entities.CartShippingRate{
			Amount: decimal.NewFromFloat(5.00),
		}, nil)

		mockCache.EXPECT().Get(ctx, gomock.Any()).Return(nil, nil).AnyTimes()

		mockVendorClient.EXPECT().CalculateTax(ctx,
			&stripeEntities.CalculateTaxRequest{
				FromAddress: stripeEntities.Address{
					City:       "La Vergne",
					State:      "TN",
					PostalCode: "37086",
					Country:    "US",
				},
				ToAddress: stripeEntities.Address{
					City:       "City",
					State:      "State",
					PostalCode: "37086",
					Country:    "US",
				},
				TaxItems: []stripeEntities.TaxItem{
					{
						Price:     decimal.NewFromFloat(10.00), // Corrected price
						Quantity:  1,
						Reference: "123456-Paperback",
						TaxCode:   stripeEntities.PaperBookTaxCode,
					},
				},
				ShippingAmount: decimal.NewFromFloat(5.00),
			},
		).Return(&stripeEntities.CalculateTaxResponse{
			Tax:      decimal.NewFromFloat(1.50),
			Currency: "USD",
		}, nil)

		mockRepo.EXPECT().UpdateCartShippingAndTaxRate(ctx, gomock.Any(), req.Body.ShippingRateUUID, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockCache.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		_, err := svc.GetTaxRate(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Error retrieving user address", func(t *testing.T) {
		svc, ctx, _, _, mockUserClient, _ := setup()
		addressUUID := uuid.New()
		req := &entities.GetTaxRateRequest{
			Body: &entities.GetTaxRateRequestBody{
				AddressUUID: addressUUID,
			},
		}

		mockUserClient.EXPECT().GetAddressByUUID(ctx, &userEntities.GetAddressRequest{
			AddressUUID: addressUUID,
		}).Return(
			nil, errors.New("database error"),
		)

		_, err := svc.GetTaxRate(ctx, req)
		assert.Error(t, err)
	})

	t.Run("Error calculating tax rate", func(t *testing.T) {
		svc, ctx, mockRepo, mockVendorClient, mockUserClient, mockCache := setup()
		req := &entities.GetTaxRateRequest{
			Body: &entities.GetTaxRateRequestBody{
				AddressUUID: uuid.New(),
			},
		}

		mockUserClient.EXPECT().GetAddressByUUID(ctx, &userEntities.GetAddressRequest{
			AddressUUID: req.Body.AddressUUID,
		}).Return(
			&userEntities.GetAddressResponse{
				Address: userEntities.UserAddress{
					City:    "City",
					State:   "State",
					Country: "US",
					ZipCode: "37086",
				},
			}, nil,
		)

		mockRepo.EXPECT().GetActiveCart(ctx, gomock.Any()).Return(&entities.Cart{
			Id:        uuid.New(),
			Status:    entities.Active,
			UpdatedAt: time.Now().Add(-time.Hour),
		}, nil)

		mockRepo.EXPECT().GetCartItems(ctx, gomock.Any()).Return([]entities.CartItem{
			{
				Id:       uuid.New(),
				Format:   entities.BookFormatPaperback,
				Price:    decimal.NewFromFloat(10.00),
				Quantity: 1,
				Length:   decimal.NewFromFloat(8.0),
				Width:    decimal.NewFromFloat(5.0),
				Height:   decimal.NewFromFloat(10.0),
				Weight:   decimal.NewFromFloat(2.0),
				Isbn:     "123456",
			},
		}, nil)

		mockRepo.EXPECT().GetShippingRate(ctx, gomock.Any()).Return(&entities.CartShippingRate{
			Amount: decimal.NewFromFloat(5.00),
		}, nil)

		mockCache.EXPECT().Get(ctx, gomock.Any()).Return(nil, nil).AnyTimes()

		mockVendorClient.EXPECT().CalculateTax(ctx,
			&stripeEntities.CalculateTaxRequest{
				FromAddress: stripeEntities.Address{
					City:       "La Vergne",
					State:      "TN",
					PostalCode: "37086",
					Country:    "US",
				},
				ToAddress: stripeEntities.Address{
					City:       "City",
					State:      "State",
					PostalCode: "37086",
					Country:    "US",
				},
				TaxItems: []stripeEntities.TaxItem{
					{
						Price:     decimal.NewFromFloat(10.00),
						Quantity:  1,
						Reference: "123456-Paperback",
						TaxCode:   stripeEntities.PaperBookTaxCode,
					},
				},
				ShippingAmount: decimal.NewFromFloat(5.00),
			},
		).Return(nil, errors.New("tax rate error"))

		_, err := svc.GetTaxRate(ctx, req)
		assert.Error(t, err)
	})
}

func Test_service_GetShippingRateByUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		mockVendorClient := stripeClient.NewMockClient(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockCache := cache.NewMockCache(ctrl)
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:         mockRepo,
			stripeClient: mockVendorClient,
			log:          zap.NewExample().Sugar(),
			userClient:   mockUserClient,
			cache:        mockCache,
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo := setup()

		shippingRateUUID := uuid.New()
		expectedRate := &entities.CartShippingRate{
			Id: shippingRateUUID,
		}

		mockRepo.EXPECT().GetShippingRate(ctx, shippingRateUUID).Return(expectedRate, nil)

		rate, err := svc.GetShippingRateByUUID(ctx, shippingRateUUID)
		assert.NoError(t, err)
		assert.Equal(t, expectedRate, rate)
	})

	t.Run("Error retrieving shipping rate", func(t *testing.T) {
		svc, ctx, mockRepo := setup()

		shippingRateUUID := uuid.New()
		mockRepo.EXPECT().GetShippingRate(ctx, shippingRateUUID).Return(nil, errors.New("database error"))

		_, err := svc.GetShippingRateByUUID(ctx, shippingRateUUID)
		assert.Error(t, err)
	})
}

func Test_service_GetCart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUUID := uuid.New()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockVendorClient := stripeClient.NewMockClient(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockCache := cache.NewMockCache(ctrl)
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:         mockRepo,
			stripeClient: mockVendorClient,
			log:          zap.NewExample().Sugar(),
			userClient:   mockUserClient,
			cache:        mockCache,
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		expectedCart := &entities.Cart{
			Id:        uuid.New(),
			Status:    entities.Active,
			UpdatedAt: time.Now(),
		}

		mockRepo.EXPECT().GetActiveCart(ctx, userUUID.String()).Return(expectedCart, nil)

		cart, err := svc.GetCart(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCart, cart)
	})

	t.Run("Error retrieving active cart", func(t *testing.T) {
		svc, ctx, mockRepo := setup()

		mockRepo.EXPECT().GetActiveCart(ctx, userUUID.String()).Return(nil, errors.New("database error"))

		_, err := svc.GetCart(ctx)
		assert.Error(t, err)
	})
}

func Test_service_GetSalesforceProudctsByBookIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		mockVendorClient := stripeClient.NewMockClient(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockCache := cache.NewMockCache(ctrl)
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:         mockRepo,
			stripeClient: mockVendorClient,
			log:          zap.NewExample().Sugar(),
			userClient:   mockUserClient,
			cache:        mockCache,
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo := setup()

		bookIDs := []string{"book1", "book2"}
		expectedProducts := []entities.SalesforceProduct{
			{BookID: "book1", SalesforceProductID: "sf1"},
			{BookID: "book2", SalesforceProductID: "sf2"},
		}

		mockRepo.EXPECT().GetSalesforceProudctsByBookIDs(ctx, bookIDs).Return(expectedProducts, nil)

		products, err := svc.GetSalesforceProudctsByBookIDs(ctx, bookIDs)
		assert.NoError(t, err)
		assert.Equal(t, expectedProducts, products)
	})

	t.Run("Error retrieving salesforce products", func(t *testing.T) {
		svc, ctx, mockRepo := setup()

		bookIDs := []string{"book1", "book2"}
		mockRepo.EXPECT().GetSalesforceProudctsByBookIDs(ctx, bookIDs).Return(nil, errors.New("database error"))

		_, err := svc.GetSalesforceProudctsByBookIDs(ctx, bookIDs)
		assert.Error(t, err)
	})
}
