package service

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

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

	// Taxes call with shipping amount 10, returns tax in minor units
	d.mockTaxes.EXPECT().
		CalculateTax(ctx, gomock.Any()).
		Do(func(_ context.Context, req *taxesEntities.CalculateTaxRequest) {
			assert.True(t, req.ShippingAmount.Equal(decimal.NewFromInt(10)))
		}).
		Return(&taxesEntities.CalculateTaxResponse{
			Tax:         decimal.NewFromInt(8500),  // 85.00
			TotalAmount: decimal.NewFromInt(18000), // 180.00
			Currency:    "USD",
			Breakdown:   sharedJson.JSON([]byte(`{"ok":true}`)),
		}, nil)

	// Update cart tax after dividing by 100 internally
	d.mockRepo.EXPECT().
		UpdateCartTaxRate(ctx, cartID.String(), decimal.NewFromFloat(85.00), "USD", gomock.Any()).
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
	assert.True(t, resp.Tax.Equal(decimal.NewFromFloat(85.00)))
	assert.True(t, resp.Total.Equal(decimal.NewFromFloat(180.00)))
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

	d.mockTaxes.EXPECT().
		CalculateTax(ctx, gomock.Any()).
		Do(func(_ context.Context, req *taxesEntities.CalculateTaxRequest) {
			assert.True(t, req.ShippingAmount.Equal(decimal.NewFromInt(12)))
		}).
		Return(&taxesEntities.CalculateTaxResponse{
			Tax:         decimal.NewFromInt(5000),  // 50.00
			TotalAmount: decimal.NewFromInt(16000), // 160.00
			Currency:    "USD",
		}, nil)

	// Update cart tax
	d.mockRepo.EXPECT().
		UpdateCartTaxRate(ctx, cartID.String(), decimal.NewFromFloat(50.00), "USD", gomock.Any()).
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
	assert.True(t, resp.Tax.Equal(decimal.NewFromFloat(50.00)))
	assert.True(t, resp.Total.Equal(decimal.NewFromFloat(160.00)))
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

	d.mockTaxes.EXPECT().
		CalculateTax(ctx, gomock.Any()).
		Do(func(_ context.Context, req *taxesEntities.CalculateTaxRequest) {
			assert.True(t, req.ShippingAmount.Equal(decimal.NewFromInt(5)))
		}).
		Return(&taxesEntities.CalculateTaxResponse{
			Tax:         decimal.NewFromInt(5000),  // 50.00
			TotalAmount: decimal.NewFromInt(15500), // 155.00
			Currency:    "USD",
		}, nil)

	// Update cart tax
	d.mockRepo.EXPECT().
		UpdateCartTaxRate(ctx, cartID.String(), decimal.NewFromFloat(50.00), "USD", gomock.Any()).
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
	assert.True(t, resp.Tax.Equal(decimal.NewFromFloat(50.00)))
	assert.True(t, resp.Total.Equal(decimal.NewFromFloat(155.00)))
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
		Total:        decimal.NewFromFloat(45.67),
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

	// Taxes client with zero shipping
	d.mockTaxes.EXPECT().
		CalculateTax(ctx, gomock.Any()).
		Do(func(_ context.Context, req *taxesEntities.CalculateTaxRequest) {
			assert.True(t, req.ShippingAmount.Equal(decimal.Zero))
		}).
		Return(&taxesEntities.CalculateTaxResponse{
			Tax:         decimal.NewFromInt(1000),
			TotalAmount: decimal.NewFromInt(7000),
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
	assert.True(t, resp.Total.Equal(decimal.NewFromInt(70)))
	assert.True(t, resp.ShippingRate.Equal(decimal.Zero))
}
