package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	addressclient "github.com/nurdsoft/nurd-commerce-core/internal/address/addressclient"
	addressEntities "github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	repository "github.com/nurdsoft/nurd-commerce-core/internal/cart/repository"
	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	sharedJson "github.com/nurdsoft/nurd-commerce-core/shared/json"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	taxes "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes"
	taxesEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/entities"
	taxesProvider "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/providers"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type testDeps struct {
	mockRepo    *repository.MockRepository
	mockTaxes   *taxes.MockClient
	mockCache   *cache.MockCache
	mockAddress *addressclient.MockClient
}

func newServiceForTest(t *testing.T) (*service, *testDeps) {
	ctrl := gomock.NewController(t)

	deps := &testDeps{
		mockRepo:    repository.NewMockRepository(ctrl),
		mockTaxes:   taxes.NewMockClient(ctrl),
		mockCache:   cache.NewMockCache(ctrl),
		mockAddress: addressclient.NewMockClient(ctrl),
	}

	logger, _ := zap.NewDevelopment()
	svc := &service{
		repo:           deps.mockRepo,
		log:            logger.Sugar(),
		shippingClient: nil,
		taxesClient:    deps.mockTaxes,
		cache:          deps.mockCache,
		productClient:  nil,
		addressClient:  deps.mockAddress,
	}

	return svc, deps
}

func TestGetTaxRate_WithOrderLevelShippingRate_UpdatesCartAndReturnsResponse(t *testing.T) {
	s, d := newServiceForTest(t)

	customerID := uuid.New().String()
	addressID := uuid.New()
	cartID := uuid.New()
	shippingRateID := uuid.New()

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID)

	d.mockAddress.EXPECT().
		GetAddress(ctx, &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{StateCode: "NY", CountryCode: "US", PostalCode: "10001"}, nil)

	// Active cart and items
	d.mockRepo.EXPECT().GetActiveCart(ctx, customerID).Return(&entities.Cart{Id: cartID}, nil)
	items := []entities.CartItemDetail{
		{ID: uuid.New(), CartID: cartID, SKU: "SKU1", Quantity: 1, Price: decimal.NewFromInt(50)},
		{ID: uuid.New(), CartID: cartID, SKU: "SKU2", Quantity: 2, Price: decimal.NewFromInt(20)},
	}
	d.mockRepo.EXPECT().GetCartItems(ctx, cartID.String()).Return(items, nil)

	// Cache miss
	expectedKey := getTaxRateCacheKey(addressID.String(), customerID, cartID.String(), shippingRateID.String())
	d.mockCache.EXPECT().Get(ctx, expectedKey).Return(nil, assert.AnError)

	// Provided order-level shipping rate and set on each item
	d.mockRepo.EXPECT().GetShippingRate(ctx, shippingRateID).
		Return(&entities.CartShippingRate{Id: shippingRateID, Amount: decimal.NewFromInt(10)}, nil)
	d.mockRepo.EXPECT().SetCartItemShippingRate(ctx, items[0].ID, shippingRateID).Return(nil)
	d.mockRepo.EXPECT().SetCartItemShippingRate(ctx, items[1].ID, shippingRateID).Return(nil)

	d.mockTaxes.EXPECT().GetProvider().Return(taxesProvider.ProviderTaxJar).Times(4) // 2 times per item

	// Taxes call with shipping amount 10, returns tax in minor units
	d.mockTaxes.EXPECT().
		CalculateTax(ctx, gomock.Any()).
		Do(func(_ context.Context, req *taxesEntities.CalculateTaxRequest) {
			assert.True(t, req.ShippingAmount.Equal(decimal.NewFromInt(10)))
		}).
		Return(&taxesEntities.CalculateTaxResponse{
			Tax:         decimal.NewFromFloat(8.50),
			TotalAmount: decimal.NewFromFloat(100.00),
			Currency:    "USD",
			Breakdown:   sharedJson.JSON([]byte(`{"ok":true}`)),
		}, nil)

	// Update cart tax after dividing by 100 internally
	d.mockRepo.EXPECT().
		UpdateCartTaxRate(ctx, cartID.String(), decimal.NewFromFloat(8.50), "USD", gomock.Any()).
		Return(nil)

	// Cache set
	d.mockCache.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	req := &entities.GetTaxRateRequest{
		Body: &entities.GetTaxRateRequestBody{
			AddressID:      addressID,
			ShippingRateID: &shippingRateID,
		},
	}

	resp, err := s.GetTaxRate(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Subtotal.Equal(decimal.NewFromInt(90)))
	assert.True(t, resp.Tax.Equal(decimal.NewFromFloat(8.50)))
	assert.True(t, resp.Total.Equal(decimal.NewFromFloat(100.00)))
	assert.True(t, resp.ShippingRate.Equal(decimal.NewFromInt(10)))
}

func TestGetTaxRate_PerItemShipping_SumsUniqueRates(t *testing.T) {
	s, d := newServiceForTest(t)

	customerID := uuid.New().String()
	addressID := uuid.New()
	cartID := uuid.New()
	rateA := uuid.New()
	rateB := uuid.New()

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID)

	d.mockAddress.EXPECT().
		GetAddress(ctx, &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{StateCode: "CA", CountryCode: "US", PostalCode: "90000"}, nil)

	d.mockRepo.EXPECT().GetActiveCart(ctx, customerID).Return(&entities.Cart{Id: cartID}, nil)
	items := []entities.CartItemDetail{
		{ID: uuid.New(), CartID: cartID, SKU: "A", Quantity: 1, Price: decimal.NewFromInt(60), ShippingRateID: &rateA},
		{ID: uuid.New(), CartID: cartID, SKU: "B", Quantity: 2, Price: decimal.NewFromInt(20), ShippingRateID: &rateB},
	}
	d.mockRepo.EXPECT().GetCartItems(ctx, cartID.String()).Return(items, nil)

	// Cache miss
	sortedShippingRateIDs := []string{rateA.String(), rateB.String()}
	sort.Strings(sortedShippingRateIDs)
	expectedKey := getTaxRateCacheKey(addressID.String(), customerID, cartID.String(), strings.Join(sortedShippingRateIDs, ","))
	d.mockCache.EXPECT().Get(ctx, expectedKey).Return(nil, assert.AnError)

	// Two unique rates: 5 + 7 = 12
	d.mockRepo.EXPECT().
		GetShippingRate(ctx, rateA).
		Return(&entities.CartShippingRate{Id: rateA, Amount: decimal.NewFromInt(5)}, nil)
	d.mockRepo.EXPECT().
		GetShippingRate(ctx, rateB).
		Return(&entities.CartShippingRate{Id: rateB, Amount: decimal.NewFromInt(7)}, nil)

	d.mockTaxes.EXPECT().GetProvider().Return(taxesProvider.ProviderTaxJar).Times(4) // 2 times per item

	d.mockTaxes.EXPECT().
		CalculateTax(ctx, gomock.Any()).
		Do(func(_ context.Context, req *taxesEntities.CalculateTaxRequest) {
			assert.True(t, req.ShippingAmount.Equal(decimal.NewFromInt(12)))
		}).
		Return(&taxesEntities.CalculateTaxResponse{
			Tax:         decimal.NewFromFloat(5.00),
			TotalAmount: decimal.NewFromFloat(100.00),
			Currency:    "USD",
		}, nil)

	// Update cart tax
	d.mockRepo.EXPECT().
		UpdateCartTaxRate(ctx, cartID.String(), decimal.NewFromFloat(5.00), "USD", gomock.Any()).
		Return(nil)

	// Cache set
	d.mockCache.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	req := &entities.GetTaxRateRequest{
		Body: &entities.GetTaxRateRequestBody{
			AddressID: addressID,
		},
	}
	resp, err := s.GetTaxRate(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Subtotal.Equal(decimal.NewFromInt(100)))
	assert.True(t, resp.Tax.Equal(decimal.NewFromFloat(5.00)))
	assert.True(t, resp.Total.Equal(decimal.NewFromFloat(100.00)))
	assert.True(t, resp.ShippingRate.Equal(decimal.NewFromInt(12)))
}

func TestGetTaxRate_PerItemShipping_SameRateForAllItems(t *testing.T) {
	s, d := newServiceForTest(t)

	customerID := uuid.New().String()
	addressID := uuid.New()
	cartID := uuid.New()
	rate := uuid.New()

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID)

	d.mockAddress.EXPECT().
		GetAddress(ctx, &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{StateCode: "CA", CountryCode: "US", PostalCode: "90000"}, nil)

	d.mockRepo.EXPECT().GetActiveCart(ctx, customerID).Return(&entities.Cart{Id: cartID}, nil)
	items := []entities.CartItemDetail{
		{ID: uuid.New(), CartID: cartID, SKU: "A", Quantity: 1, Price: decimal.NewFromInt(60), ShippingRateID: &rate},
		{ID: uuid.New(), CartID: cartID, SKU: "B", Quantity: 2, Price: decimal.NewFromInt(20), ShippingRateID: &rate},
	}
	d.mockRepo.EXPECT().GetCartItems(ctx, cartID.String()).Return(items, nil)

	// Cache miss
	expectedKey := getTaxRateCacheKey(addressID.String(), customerID, cartID.String(), rate.String())
	d.mockCache.EXPECT().Get(ctx, expectedKey).Return(nil, assert.AnError)

	d.mockRepo.EXPECT().
		GetShippingRate(ctx, rate).
		Return(&entities.CartShippingRate{Id: rate, Amount: decimal.NewFromInt(5)}, nil)

	d.mockTaxes.EXPECT().GetProvider().Return(taxesProvider.ProviderTaxJar).Times(4) // 2 times per item

	d.mockTaxes.EXPECT().
		CalculateTax(ctx, gomock.Any()).
		Do(func(_ context.Context, req *taxesEntities.CalculateTaxRequest) {
			assert.True(t, req.ShippingAmount.Equal(decimal.NewFromInt(5)))
		}).
		Return(&taxesEntities.CalculateTaxResponse{
			Tax:         decimal.NewFromFloat(5.00),
			TotalAmount: decimal.NewFromFloat(100.00),
			Currency:    "USD",
		}, nil)

	// Update cart tax
	d.mockRepo.EXPECT().
		UpdateCartTaxRate(ctx, cartID.String(), decimal.NewFromFloat(5.00), "USD", gomock.Any()).
		Return(nil)

	// Cache set
	d.mockCache.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	req := &entities.GetTaxRateRequest{
		Body: &entities.GetTaxRateRequestBody{
			AddressID: addressID,
		},
	}
	resp, err := s.GetTaxRate(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Subtotal.Equal(decimal.NewFromInt(100)))
	assert.True(t, resp.Tax.Equal(decimal.NewFromFloat(5.00)))
	assert.True(t, resp.Total.Equal(decimal.NewFromFloat(100.00)))
	assert.True(t, resp.ShippingRate.Equal(decimal.NewFromInt(5)))
}

func TestGetTaxRate_CacheHit(t *testing.T) {
	s, d := newServiceForTest(t)

	customerID := uuid.New().String()
	addressID := uuid.New()
	shippingRateID := uuid.New()
	cartID := uuid.New()

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID)

	// Address fetched first
	d.mockAddress.EXPECT().
		GetAddress(ctx, &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{StateCode: "CA", CountryCode: "US", PostalCode: "90000"}, nil)

	// Active cart/items fetched to build cache key
	d.mockRepo.EXPECT().GetActiveCart(ctx, customerID).Return(&entities.Cart{Id: cartID}, nil)
	d.mockRepo.EXPECT().
		GetCartItems(ctx, cartID.String()).
		Return([]entities.CartItemDetail{
			{
				ID:             uuid.New(),
				CartID:         cartID,
				SKU:            "X",
				Quantity:       1,
				Price:          decimal.NewFromInt(10),
				ShippingRateID: &shippingRateID,
			},
		}, nil)

	d.mockRepo.EXPECT().
		GetShippingRate(ctx, shippingRateID).
		Return(&entities.CartShippingRate{Id: shippingRateID, Amount: decimal.NewFromInt(444)}, nil)

	cached := entities.GetTaxRateResponse{
		Tax:          decimal.NewFromFloat(1.23),
		Total:        decimal.NewFromFloat(44.44),
		Subtotal:     decimal.NewFromFloat(40.00),
		ShippingRate: decimal.NewFromFloat(4.44),
		Currency:     "USD",
	}
	b, _ := json.Marshal(cached)
	expectedKey := getTaxRateCacheKey(addressID.String(), customerID, cartID.String(), shippingRateID.String())
	d.mockCache.EXPECT().Get(ctx, expectedKey).Return(b, nil)

	req := &entities.GetTaxRateRequest{
		Body: &entities.GetTaxRateRequestBody{
			AddressID: addressID,
		},
	}
	resp, err := s.GetTaxRate(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Tax.Equal(cached.Tax))
	assert.True(t, resp.Total.Equal(cached.Total))
	assert.True(t, resp.Subtotal.Equal(cached.Subtotal))
	assert.True(t, resp.ShippingRate.Equal(cached.ShippingRate))
}

func TestGetTaxRate_NoShippingRates_ShippingZero(t *testing.T) {
	s, d := newServiceForTest(t)

	customerID := uuid.New().String()
	addressID := uuid.New()
	cartID := uuid.New()

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID)

	d.mockAddress.EXPECT().
		GetAddress(ctx, &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{StateCode: "CA", CountryCode: "US", PostalCode: "90000"}, nil)

	d.mockRepo.EXPECT().GetActiveCart(ctx, customerID).Return(&entities.Cart{Id: cartID}, nil)
	d.mockRepo.EXPECT().GetCartItems(ctx, cartID.String()).Return([]entities.CartItemDetail{
		{ID: uuid.New(), CartID: cartID, SKU: "A", Quantity: 2, Price: decimal.NewFromInt(25)},
		{ID: uuid.New(), CartID: cartID, SKU: "B", Quantity: 1, Price: decimal.NewFromInt(15)},
	}, nil)

	// Cache miss
	d.mockCache.EXPECT().Get(ctx, gomock.Any()).Return(nil, assert.AnError)

	d.mockTaxes.EXPECT().GetProvider().Return(taxesProvider.ProviderTaxJar).Times(4) // 2 times per item

	// Taxes client with zero shipping
	d.mockTaxes.EXPECT().
		CalculateTax(ctx, gomock.Any()).
		Do(func(_ context.Context, req *taxesEntities.CalculateTaxRequest) {
			assert.True(t, req.ShippingAmount.Equal(decimal.Zero))
		}).
		Return(&taxesEntities.CalculateTaxResponse{
			Tax:         decimal.NewFromFloat(10.00),
			TotalAmount: decimal.NewFromFloat(75.00),
			Currency:    "USD",
		}, nil)

	// Update tax
	d.mockRepo.EXPECT().
		UpdateCartTaxRate(ctx, cartID.String(), decimal.NewFromFloat(10.00), "USD", gomock.Any()).
		Return(nil)

	// Cache set
	d.mockCache.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	req := &entities.GetTaxRateRequest{
		Body: &entities.GetTaxRateRequestBody{
			AddressID: addressID,
		},
	}
	resp, err := s.GetTaxRate(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Total.Equal(decimal.NewFromFloat(75.00)))
	assert.True(t, resp.ShippingRate.Equal(decimal.Zero))
}

func TestSetCartItemShippingRate_Success(t *testing.T) {
	s, d := newServiceForTest(t)

	customerID := uuid.New().String()
	shippingRateID := uuid.New()
	cartItemID := uuid.New()

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID)

	d.mockRepo.EXPECT().
		GetCartItemByID(ctx, cartItemID).
		Return(&entities.CartItem{ID: cartItemID, CartID: uuid.New()}, nil)

	d.mockRepo.EXPECT().GetShippingRate(ctx, shippingRateID).
		Return(&entities.CartShippingRate{Id: shippingRateID, Amount: decimal.NewFromInt(10)}, nil)

	d.mockRepo.EXPECT().SetCartItemShippingRate(ctx, cartItemID, shippingRateID).Return(nil)

	deleteCacheCallDone := make(chan struct{})
	// Check that the cache was cleared
	d.mockCache.EXPECT().
		DeleteByPattern(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, pattern string) error {
			defer close(deleteCacheCallDone)
			assert.Equal(t, fmt.Sprintf("^tax_rate_[^_]+_%s_", customerID), pattern)
			return nil
		})

	req := &entities.SetCartItemShippingRateRequest{
		Body: &entities.SetCartItemShippingRateRequestBody{
			CartItemID:     cartItemID,
			ShippingRateID: shippingRateID,
		},
	}

	err := s.SetCartItemShippingRate(ctx, req)

	assert.NoError(t, err)
	<-deleteCacheCallDone
}

func TestCreateCartShippingRates_Success(t *testing.T) {
	s, d := newServiceForTest(t)

	customerID := uuid.New().String()
	cartID := uuid.New()
	addressID := uuid.New()
	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID)

	d.mockAddress.EXPECT().
		GetAddress(ctx, &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{StateCode: "CA", CountryCode: "US", PostalCode: "90000"}, nil)

	d.mockRepo.EXPECT().GetActiveCart(ctx, customerID).Return(&entities.Cart{Id: cartID}, nil)

	estimatedDeliveryDate := time.Now().Add(time.Hour * 24 * 3)
	estimatedDeliveryDateExpress := time.Now().Add(time.Hour * 24 * 1)
	expectedRates := []entities.CartShippingRate{
		{
			Amount:                decimal.NewFromInt(10),
			Currency:              "USD",
			CarrierName:           "UPS",
			CarrierCode:           "UPS",
			ServiceType:           "Standard",
			ServiceCode:           "123456",
			EstimatedDeliveryDate: estimatedDeliveryDate,
			BusinessDaysInTransit: "3",
		},
		{
			Amount:                decimal.NewFromInt(15),
			Currency:              "USD",
			CarrierName:           "FedEx",
			CarrierCode:           "FedEx",
			ServiceType:           "Express",
			ServiceCode:           "123457",
			EstimatedDeliveryDate: estimatedDeliveryDateExpress,
			BusinessDaysInTransit: "3",
		},
	}

	d.mockRepo.EXPECT().CreateCartShippingRates(ctx, gomock.Any()).Return(nil)

	req := &entities.CreateCartShippingRatesRequest{
		Body: &entities.CreateCartShippingRatesRequestBody{
			AddressID: addressID,
			CartShippingRates: []entities.CartShippingRateRequest{
				{
					Amount:                decimal.NewFromInt(10),
					Currency:              "USD",
					CarrierName:           "UPS",
					CarrierCode:           "UPS",
					ServiceType:           "Standard",
					ServiceCode:           "123456",
					EstimatedDeliveryDate: estimatedDeliveryDate,
					BusinessDaysInTransit: "3",
				},
				{
					Amount:                decimal.NewFromInt(15),
					Currency:              "USD",
					CarrierName:           "FedEx",
					CarrierCode:           "FedEx",
					ServiceType:           "Express",
					ServiceCode:           "123457",
					EstimatedDeliveryDate: estimatedDeliveryDateExpress,
					BusinessDaysInTransit: "3",
				},
			},
		},
	}

	resp, err := s.CreateCartShippingRates(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, len(resp.Rates), 2)
	assert.NotEmpty(t, resp.Rates[0].Id)
	assert.Equal(t, resp.Rates[0].Amount, expectedRates[0].Amount)
	assert.Equal(t, resp.Rates[0].Currency, expectedRates[0].Currency)
	assert.Equal(t, resp.Rates[0].CarrierName, expectedRates[0].CarrierName)
	assert.Equal(t, resp.Rates[0].CarrierCode, expectedRates[0].CarrierCode)
	assert.Equal(t, resp.Rates[0].ServiceType, expectedRates[0].ServiceType)
	assert.Equal(t, resp.Rates[0].ServiceCode, expectedRates[0].ServiceCode)
	assert.Equal(t, resp.Rates[0].EstimatedDeliveryDate, expectedRates[0].EstimatedDeliveryDate)
	assert.Equal(t, resp.Rates[0].BusinessDaysInTransit, expectedRates[0].BusinessDaysInTransit)

	assert.NotEmpty(t, resp.Rates[1].Id)
	assert.Equal(t, resp.Rates[1].Amount, expectedRates[1].Amount)
	assert.Equal(t, resp.Rates[1].Currency, expectedRates[1].Currency)
	assert.Equal(t, resp.Rates[1].CarrierName, expectedRates[1].CarrierName)
	assert.Equal(t, resp.Rates[1].CarrierCode, expectedRates[1].CarrierCode)
	assert.Equal(t, resp.Rates[1].ServiceType, expectedRates[1].ServiceType)
	assert.Equal(t, resp.Rates[1].ServiceCode, expectedRates[1].ServiceCode)
	assert.Equal(t, resp.Rates[1].EstimatedDeliveryDate, expectedRates[1].EstimatedDeliveryDate)
	assert.Equal(t, resp.Rates[1].BusinessDaysInTransit, expectedRates[1].BusinessDaysInTransit)
}
