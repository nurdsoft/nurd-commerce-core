package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/addressclient"
	addressEntities "github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	customerEntities "github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/orders/errors"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/repository"
	productEntities "github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	webhookclient "github.com/nurdsoft/nurd-commerce-core/internal/webhook/client"
	webhookEntities "github.com/nurdsoft/nurd-commerce-core/internal/webhook/entities"
	wishlistEntities "github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	wishlistclient "github.com/nurdsoft/nurd-commerce-core/internal/wishlist/wishlistclient"
	sharedErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/nurdsoft/nurd-commerce-core/shared/nullable"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	inventoryEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment"
	authorizenetEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type testController struct {
	mockRepo      *repository.MockRepository
	mockCustomer  *customerclient.MockClient
	mockCart      *cartclient.MockClient
	mockPayment   *payment.MockClient
	mockWishlist  *wishlistclient.MockClient
	mockInventory *inventory.MockClient
	mockAddress   *addressclient.MockClient
	mockProduct   *productclient.MockClient
	mockWebhook   *webhookclient.MockClient
}

func setupTestController(t *testing.T) *testController {
	ctrl := gomock.NewController(t)

	return &testController{
		mockRepo:      repository.NewMockRepository(ctrl),
		mockCustomer:  customerclient.NewMockClient(ctrl),
		mockCart:      cartclient.NewMockClient(ctrl),
		mockPayment:   payment.NewMockClient(ctrl),
		mockWishlist:  wishlistclient.NewMockClient(ctrl),
		mockInventory: inventory.NewMockClient(ctrl),
		mockAddress:   addressclient.NewMockClient(ctrl),
		mockProduct:   productclient.NewMockClient(ctrl),
		mockWebhook:   webhookclient.NewMockClient(ctrl),
	}
}

func newServiceUnderTest(tc *testController) *service {
	logger, _ := zap.NewDevelopment()
	return &service{
		repo:            tc.mockRepo,
		log:             logger.Sugar(),
		customerClient:  tc.mockCustomer,
		cartClient:      tc.mockCart,
		paymentClient:   tc.mockPayment,
		wishlistClient:  tc.mockWishlist,
		inventoryClient: tc.mockInventory,
		addressClient:   tc.mockAddress,
		productClient:   tc.mockProduct,
		webhookClient:   tc.mockWebhook,
	}
}

func TestCreateOrder_WithStripe(t *testing.T) {
	tc := setupTestController(t)
	s := newServiceUnderTest(tc)

	customerID := uuid.New()
	addressID := uuid.New()
	cartID := uuid.New()
	shippingRateID := uuid.New()
	paymentMethodID := "pm_123"
	customerStripeID := "cus_123"
	expectedPaymentIntentID := "pi_123"
	expectedAddress := "123 Main St"
	expectedTotal := decimal.NewFromInt(115)

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

	tc.mockAddress.EXPECT().
		GetAddress(gomock.Any(), &addressEntities.GetAddressRequest{
			AddressID: addressID,
		}).
		Return(&addressEntities.Address{
			FullName:    "John Doe",
			Address:     expectedAddress,
			City:        nullable.StringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: nullable.StringPtr("1234567890"),
		}, nil)

	tc.mockCart.EXPECT().
		GetCart(gomock.Any()).
		Return(&cartEntities.Cart{
			Id:          cartID,
			TaxAmount:   decimal.NewFromFloat(10.0),
			TaxCurrency: "USD",
		}, nil)

	tc.mockCart.EXPECT().
		GetCartItems(gomock.Any()).
		Return(&cartEntities.GetCartItemsResponse{
			Items: []cartEntities.CartItemDetail{
				{
					ProductID:        uuid.New(),
					ProductVariantID: uuid.New(),
					SKU:              "SKU123",
					Name:             "Test Product",
					Quantity:         2,
					Price:            decimal.NewFromInt(50),
					ShippingRateID:   &shippingRateID,
				},
			},
		}, nil)

	tc.mockCart.EXPECT().
		GetShippingRateByID(gomock.Any(), shippingRateID).
		Return(&cartEntities.CartShippingRate{
			Id:                    shippingRateID,
			Amount:                decimal.NewFromInt(5),
			CarrierName:           "Test Carrier",
			CarrierCode:           "TEST",
			ServiceType:           "Standard",
			ServiceCode:           "STD",
			EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
			BusinessDaysInTransit: "2",
		}, nil)

	tc.mockCustomer.EXPECT().
		GetCustomer(gomock.Any()).
		Return(&customerEntities.Customer{
			ID:       customerID,
			StripeID: nullable.StringPtr(customerStripeID),
		}, nil)

	tc.mockPayment.EXPECT().
		GetProvider().
		Return(providers.ProviderStripe).Times(2)

	tc.mockPayment.EXPECT().
		CreatePayment(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, req stripeEntities.CreatePaymentIntentRequest) {
			assert.Equal(t, expectedTotal, req.Amount)
			assert.Equal(t, paymentMethodID, req.PaymentMethodId)
			assert.Equal(t, customerStripeID, *req.CustomerId)
		}).
		Return(providers.PaymentProviderResponse{
			ID:     expectedPaymentIntentID,
			Status: providers.PaymentStatusPending,
		}, nil)

	tc.mockRepo.EXPECT().
		OrderReferenceExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	tc.mockRepo.EXPECT().
		CreateOrder(gomock.Any(), cartID, gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ uuid.UUID, order *entities.Order, _ []*entities.OrderItem) {
			//assert order details
			assert.Equal(t, customerID, order.CustomerID)
			assert.Equal(t, cartID, order.CartID)
			assert.Equal(t, expectedAddress, order.DeliveryAddress)
			assert.Equal(t, expectedTotal, order.Total)
			assert.Equal(t, entities.Pending, order.Status)
			assert.Equal(t, expectedPaymentIntentID, *order.StripePaymentIntentID)
			assert.Equal(t, paymentMethodID, order.StripePaymentMethodID)
		}).
		Return(nil)

	notifyCallDone := make(chan struct{})
	tc.mockWebhook.EXPECT().
		NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, req *webhookEntities.NotifyOrderStatusChangeRequest) {
			assert.Equal(t, customerID.String(), req.CustomerID)
			assert.Equal(t, entities.Pending.String(), req.Status)
			close(notifyCallDone)
		}).
		Return(nil)

	salesforceCallDone := make(chan struct{})
	tc.mockInventory.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ inventoryEntities.CreateInventoryOrderRequest) {
			close(salesforceCallDone)
		}).
		Return(nil, nil)

	req := &entities.CreateOrderRequest{
		Body: &entities.CreateOrderRequestBody{
			AddressID:             addressID,
			ShippingRateID:        &shippingRateID,
			StripePaymentMethodID: paymentMethodID,
		},
	}

	resp, err := s.CreateOrder(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.OrderReference)
	// wait for async notify call to be done
	<-notifyCallDone
	<-salesforceCallDone
}

func TestCreateOrder_WithAuthorizeNet(t *testing.T) {
	tc := setupTestController(t)
	s := newServiceUnderTest(tc)

	customerID := uuid.New()
	addressID := uuid.New()
	cartID := uuid.New()
	shippingRateID := uuid.New()
	paymentNonce := "fake-nonce"
	customerAuthorizeNetID := "123456"
	expectedTransactionID := "123456"
	expectedAddress := "123 Main St"
	expectedTotal := decimal.NewFromInt(115)
	expectedBillingInfo := entities.BillingInfo{
		FirstName: "John",
		LastName:  "Doe",
		Address:   "123 Main St",
		City:      "Anytown",
		State:     "CA",
		Country:   "US",
		Zip:       "12345",
	}

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

	tc.mockAddress.EXPECT().
		GetAddress(gomock.Any(), &addressEntities.GetAddressRequest{
			AddressID: addressID,
		}).
		Return(&addressEntities.Address{
			FullName:    "John Doe",
			Address:     expectedAddress,
			City:        nullable.StringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: nullable.StringPtr("1234567890"),
		}, nil)

	tc.mockCart.EXPECT().
		GetCart(gomock.Any()).
		Return(&cartEntities.Cart{
			Id:          cartID,
			TaxAmount:   decimal.NewFromInt(10),
			TaxCurrency: "USD",
		}, nil)

	tc.mockCart.EXPECT().
		GetCartItems(gomock.Any()).
		Return(&cartEntities.GetCartItemsResponse{
			Items: []cartEntities.CartItemDetail{
				{
					ProductID:        uuid.New(),
					ProductVariantID: uuid.New(),
					SKU:              "SKU123",
					Name:             "Test Product",
					Quantity:         2,
					Price:            decimal.NewFromInt(50),
					ShippingRateID:   &shippingRateID,
				},
			},
		}, nil)

	tc.mockCart.EXPECT().
		GetShippingRateByID(gomock.Any(), shippingRateID).
		Return(&cartEntities.CartShippingRate{
			Id:                    shippingRateID,
			Amount:                decimal.NewFromInt(5),
			CarrierName:           "Test Carrier",
			CarrierCode:           "TEST",
			ServiceType:           "Standard",
			ServiceCode:           "STD",
			EstimatedDeliveryDate: time.Now().Add(24 * time.Hour),
			BusinessDaysInTransit: "2",
		}, nil)

	tc.mockCustomer.EXPECT().
		GetCustomer(gomock.Any()).
		Return(&customerEntities.Customer{
			ID:             customerID,
			AuthorizeNetID: nullable.StringPtr(customerAuthorizeNetID),
		}, nil)

	tc.mockPayment.EXPECT().
		GetProvider().
		Return(providers.ProviderAuthorizeNet).Times(2)

	tc.mockPayment.EXPECT().
		CreatePayment(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, req authorizenetEntities.CreatePaymentTransactionRequest) {
			assert.Equal(t, expectedTotal, req.Amount)
			assert.Equal(t, paymentNonce, req.PaymentNonce)
			assert.EqualValues(t, expectedBillingInfo, req.BillingInfo)
		}).
		Return(providers.PaymentProviderResponse{
			ID:     expectedTransactionID,
			Status: providers.PaymentStatusSuccess,
		}, nil)

	tc.mockRepo.EXPECT().
		OrderReferenceExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	tc.mockRepo.EXPECT().
		CreateOrder(gomock.Any(), cartID, gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ uuid.UUID, order *entities.Order, _ []*entities.OrderItem) {
			//assert order details
			assert.Equal(t, customerID, order.CustomerID)
			assert.Equal(t, cartID, order.CartID)
			assert.Equal(t, expectedAddress, order.DeliveryAddress)
			assert.Equal(t, expectedTotal, order.Total)
			assert.Equal(t, entities.PaymentSuccess, order.Status)
			assert.Equal(t, expectedTransactionID, *order.AuthorizeNetPaymentID)
			assert.Empty(t, order.StripePaymentMethodID)
		}).
		Return(nil)

	notifyCallDone := make(chan struct{})
	tc.mockWebhook.EXPECT().
		NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, req *webhookEntities.NotifyOrderStatusChangeRequest) {
			assert.Equal(t, customerID.String(), req.CustomerID)
			assert.Equal(t, entities.PaymentSuccess.String(), req.Status)
			close(notifyCallDone)
		}).
		Return(nil)

	salesforceCallDone := make(chan struct{})
	tc.mockInventory.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ inventoryEntities.CreateInventoryOrderRequest) {
			close(salesforceCallDone)
		}).
		Return(nil, nil)

	req := &entities.CreateOrderRequest{
		Body: &entities.CreateOrderRequestBody{
			AddressID:    addressID,
			PaymentNonce: paymentNonce,
			BillingInfo:  expectedBillingInfo,
		},
	}

	resp, err := s.CreateOrder(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.OrderReference)
	// wait for async notify call to be done
	<-notifyCallDone
	<-salesforceCallDone
}

func TestCreateOrder_PerItemShipping_MultipleRates_Stripe(t *testing.T) {
	tc := setupTestController(t)
	s := newServiceUnderTest(tc)

	customerID := uuid.New()
	addressID := uuid.New()
	cartID := uuid.New()
	shippingRateID1 := uuid.New()
	shippingRateID2 := uuid.New()
	paymentMethodID := "pm_123"
	customerStripeID := "cus_456"

	// subtotal: (50*1) + (30*2) = 110
	// tax: 10
	// shipping: 5 + 7 = 12
	expectedTotal := decimal.NewFromInt(132)

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

	tc.mockAddress.EXPECT().
		GetAddress(gomock.Any(), &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{
			FullName:    "John Doe",
			Address:     "123 Main St",
			City:        nullable.StringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: nullable.StringPtr("1234567890"),
		}, nil)

	tc.mockCart.EXPECT().
		GetCart(gomock.Any()).
		Return(&cartEntities.Cart{Id: cartID, TaxAmount: decimal.NewFromInt(10), TaxCurrency: "USD"}, nil)

	tc.mockCart.EXPECT().
		GetCartItems(gomock.Any()).
		Return(&cartEntities.GetCartItemsResponse{Items: []cartEntities.CartItemDetail{
			{
				ProductID:        uuid.New(),
				ProductVariantID: uuid.New(),
				SKU:              "SKU-A",
				Name:             "Product A",
				Quantity:         1,
				Price:            decimal.NewFromInt(50),
				ShippingRateID:   &shippingRateID1,
			},
			{
				ProductID:        uuid.New(),
				ProductVariantID: uuid.New(),
				SKU:              "SKU-B",
				Name:             "Product B",
				Quantity:         2,
				Price:            decimal.NewFromInt(30),
				ShippingRateID:   &shippingRateID2,
			},
		}}, nil)

	tc.mockCart.EXPECT().
		GetShippingRateByID(gomock.Any(), shippingRateID1).
		Return(&cartEntities.CartShippingRate{
			Id:          shippingRateID1,
			Amount:      decimal.NewFromInt(5),
			CarrierName: "Carrier 1",
			CarrierCode: "C1",
			ServiceType: "Standard",
			ServiceCode: "STD",
		}, nil)

	tc.mockCart.EXPECT().
		GetShippingRateByID(gomock.Any(), shippingRateID2).
		Return(&cartEntities.CartShippingRate{
			Id:          shippingRateID2,
			Amount:      decimal.NewFromInt(7),
			CarrierName: "Carrier 2",
			CarrierCode: "C2",
			ServiceType: "Express",
			ServiceCode: "EXP",
		}, nil)

	tc.mockCustomer.EXPECT().
		GetCustomer(gomock.Any()).
		Return(&customerEntities.Customer{ID: customerID, StripeID: nullable.StringPtr(customerStripeID)}, nil)

	tc.mockPayment.EXPECT().
		GetProvider().
		Return(providers.ProviderStripe).Times(2)

	tc.mockPayment.EXPECT().
		CreatePayment(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, req stripeEntities.CreatePaymentIntentRequest) {
			assert.Equal(t, expectedTotal, req.Amount)
			assert.Equal(t, paymentMethodID, req.PaymentMethodId)
			assert.Equal(t, customerStripeID, *req.CustomerId)
		}).
		Return(providers.PaymentProviderResponse{ID: "pi_multi", Status: providers.PaymentStatusPending}, nil)

	tc.mockRepo.EXPECT().
		OrderReferenceExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	tc.mockRepo.EXPECT().
		CreateOrder(gomock.Any(), cartID, gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ uuid.UUID, order *entities.Order, items []*entities.OrderItem) {
			assert.Equal(t, expectedTotal, order.Total)
			// Order-level shipping rate is set to total, but carrier fields remain nil when not provided by request
			assert.Equal(t, decimal.NewFromInt(12), *order.ShippingRate)
			assert.Empty(t, order.ShippingCarrierName)
			assert.Empty(t, order.ShippingCarrierCode)
			assert.Empty(t, order.ShippingEstimatedDeliveryDate)
			assert.Empty(t, order.ShippingBusinessDaysInTransit)
			assert.Empty(t, order.ShippingServiceType)
			assert.Empty(t, order.ShippingServiceCode)

			// items should have shipping rate id and amount set
			assert.Len(t, items, 2)
			found1 := false
			found2 := false
			for _, it := range items {
				if it.ShippingRateID != nil && *it.ShippingRateID == shippingRateID1 {
					found1 = true
					assert.Equal(t, decimal.NewFromInt(5), *it.ShippingRate)
				}
				if it.ShippingRateID != nil && *it.ShippingRateID == shippingRateID2 {
					found2 = true
					assert.Equal(t, decimal.NewFromInt(7), *it.ShippingRate)
				}
			}
			assert.True(t, found1)
			assert.True(t, found2)
		}).
		Return(nil)

	notifyDone := make(chan struct{})
	tc.mockWebhook.EXPECT().
		NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ *webhookEntities.NotifyOrderStatusChangeRequest) { close(notifyDone) }).
		Return(nil)

	salesforceCallDone := make(chan struct{})
	tc.mockInventory.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ inventoryEntities.CreateInventoryOrderRequest) { close(salesforceCallDone) }).
		Return(nil, nil)

	req := &entities.CreateOrderRequest{Body: &entities.CreateOrderRequestBody{AddressID: addressID, StripePaymentMethodID: paymentMethodID}}
	resp, err := s.CreateOrder(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	<-notifyDone
	<-salesforceCallDone
}

func TestCreateOrder_BackCompat_OrderLevelShipping_SetsOrderFields(t *testing.T) {
	tc := setupTestController(t)
	s := newServiceUnderTest(tc)

	customerID := uuid.New()
	addressID := uuid.New()
	cartID := uuid.New()
	shippingRateID := uuid.New()
	paymentMethodID := "pm_789"
	customerStripeID := "cus_789"

	// subtotal: 40 + 40 = 80; tax: 10; shipping (single, shared) = 5; total = 95
	expectedTotal := decimal.NewFromInt(95)

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

	tc.mockAddress.EXPECT().
		GetAddress(gomock.Any(), &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{
			FullName:    "John Doe",
			Address:     "123 Main St",
			City:        nullable.StringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: nullable.StringPtr("1234567890"),
		}, nil)

	tc.mockCart.EXPECT().
		GetCart(gomock.Any()).
		Return(&cartEntities.Cart{Id: cartID, TaxAmount: decimal.NewFromInt(10), TaxCurrency: "USD"}, nil)

	tc.mockCart.EXPECT().
		GetCartItems(gomock.Any()).
		Return(&cartEntities.GetCartItemsResponse{Items: []cartEntities.CartItemDetail{
			{
				ProductID:        uuid.New(),
				ProductVariantID: uuid.New(),
				SKU:              "SKU-X",
				Name:             "X",
				Quantity:         1,
				Price:            decimal.NewFromInt(40),
				ShippingRateID:   &shippingRateID,
			},
			{
				ProductID:        uuid.New(),
				ProductVariantID: uuid.New(),
				SKU:              "SKU-Y",
				Name:             "Y",
				Quantity:         1,
				Price:            decimal.NewFromInt(40),
				ShippingRateID:   &shippingRateID,
			},
		}}, nil)

	// The rate is looked up per item (twice)
	tc.mockCart.EXPECT().
		GetShippingRateByID(gomock.Any(), shippingRateID).
		Return(&cartEntities.CartShippingRate{
			Id:          shippingRateID,
			Amount:      decimal.NewFromInt(5),
			CarrierName: "Carrier",
			CarrierCode: "CARR",
			ServiceType: "Ground",
			ServiceCode: "GRD",
		}, nil).Times(2)

	tc.mockCustomer.EXPECT().
		GetCustomer(gomock.Any()).
		Return(&customerEntities.Customer{ID: customerID, StripeID: nullable.StringPtr(customerStripeID)}, nil)

	tc.mockPayment.EXPECT().
		GetProvider().
		Return(providers.ProviderStripe).Times(2)

	tc.mockPayment.EXPECT().
		CreatePayment(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, req stripeEntities.CreatePaymentIntentRequest) {
			assert.Equal(t, expectedTotal, req.Amount)
			assert.Equal(t, paymentMethodID, req.PaymentMethodId)
			assert.Equal(t, customerStripeID, *req.CustomerId)
		}).
		Return(providers.PaymentProviderResponse{ID: "pi_backcompat", Status: providers.PaymentStatusPending}, nil)

	tc.mockRepo.EXPECT().
		OrderReferenceExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	tc.mockRepo.EXPECT().
		CreateOrder(gomock.Any(), cartID, gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ uuid.UUID, order *entities.Order, items []*entities.OrderItem) {
			assert.Equal(t, expectedTotal, order.Total)
			assert.Equal(t, decimal.NewFromInt(5), *order.ShippingRate)
			// Order-level shipping fields should be set when request includes ShippingRateID
			assert.Equal(t, "Carrier", *order.ShippingCarrierName)
			assert.Equal(t, "CARR", *order.ShippingCarrierCode)
			assert.Equal(t, "Ground", *order.ShippingServiceType)
			assert.Equal(t, "GRD", *order.ShippingServiceCode)
			assert.Len(t, items, 2)
			for _, it := range items {
				assert.Equal(t, shippingRateID, *it.ShippingRateID)
				assert.Equal(t, decimal.NewFromInt(5), *it.ShippingRate)
			}
		}).
		Return(nil)

	notifyDone := make(chan struct{})
	tc.mockWebhook.EXPECT().
		NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ *webhookEntities.NotifyOrderStatusChangeRequest) { close(notifyDone) }).
		Return(nil)

	salesforceCallDone := make(chan struct{})
	tc.mockInventory.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ inventoryEntities.CreateInventoryOrderRequest) { close(salesforceCallDone) }).
		Return(nil, nil)

	req := &entities.CreateOrderRequest{Body: &entities.CreateOrderRequestBody{AddressID: addressID, ShippingRateID: &shippingRateID, StripePaymentMethodID: paymentMethodID}}
	resp, err := s.CreateOrder(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	<-notifyDone
	<-salesforceCallDone
}

func TestCreateOrder_Error_MismatchedOrderLevelShippingRate(t *testing.T) {
	tc := setupTestController(t)
	s := newServiceUnderTest(tc)

	customerID := uuid.New()
	addressID := uuid.New()
	cartID := uuid.New()
	itemRateID := uuid.New()
	orderRateID := uuid.New() // different

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

	tc.mockAddress.EXPECT().
		GetAddress(gomock.Any(), &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{
			FullName:    "John Doe",
			Address:     "123 Main St",
			City:        nullable.StringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: nullable.StringPtr("1234567890"),
		}, nil)

	tc.mockCart.EXPECT().
		GetCart(gomock.Any()).
		Return(&cartEntities.Cart{Id: cartID, TaxAmount: decimal.NewFromInt(10), TaxCurrency: "USD"}, nil)

	tc.mockCart.EXPECT().
		GetCartItems(gomock.Any()).
		Return(&cartEntities.GetCartItemsResponse{Items: []cartEntities.CartItemDetail{{
			ProductID:        uuid.New(),
			ProductVariantID: uuid.New(),
			SKU:              "SKU-1",
			Name:             "One",
			Quantity:         1,
			Price:            decimal.NewFromInt(10),
			ShippingRateID:   &itemRateID,
		},
		}}, nil)

	tc.mockCart.EXPECT().
		GetShippingRateByID(gomock.Any(), itemRateID).
		Return(&cartEntities.CartShippingRate{Id: itemRateID, Amount: decimal.NewFromInt(5)}, nil)

	req := &entities.CreateOrderRequest{Body: &entities.CreateOrderRequestBody{AddressID: addressID, ShippingRateID: &orderRateID}}
	resp, err := s.CreateOrder(ctx, req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, moduleErrors.NewAPIError("ORDER_ERROR_CREATING").Error())
}

func TestCreateOrder_DuplicateShippingRateIDs_SummedOnce(t *testing.T) {
	tc := setupTestController(t)
	s := newServiceUnderTest(tc)

	customerID := uuid.New()
	addressID := uuid.New()
	cartID := uuid.New()
	sharedRateID := uuid.New()
	paymentMethodID := "pm_dup"
	customerStripeID := "cus_dup"

	// subtotal: 20 + 20 = 40; tax: 10; shipping: 5 (only once) => total 55
	expectedTotal := decimal.NewFromInt(55)

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

	tc.mockAddress.EXPECT().
		GetAddress(gomock.Any(), &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{
			FullName:    "John Doe",
			Address:     "123 Main St",
			City:        nullable.StringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: nullable.StringPtr("1234567890"),
		}, nil)

	tc.mockCart.EXPECT().
		GetCart(gomock.Any()).
		Return(&cartEntities.Cart{Id: cartID, TaxAmount: decimal.NewFromInt(10), TaxCurrency: "USD"}, nil)

	tc.mockCart.EXPECT().
		GetCartItems(gomock.Any()).
		Return(&cartEntities.GetCartItemsResponse{Items: []cartEntities.CartItemDetail{
			{
				ProductID:        uuid.New(),
				ProductVariantID: uuid.New(),
				SKU:              "SKU-1",
				Name:             "One",
				Quantity:         1,
				Price:            decimal.NewFromInt(20),
				ShippingRateID:   &sharedRateID,
			},
			{
				ProductID:        uuid.New(),
				ProductVariantID: uuid.New(),
				SKU:              "SKU-2",
				Name:             "Two",
				Quantity:         1,
				Price:            decimal.NewFromInt(20),
				ShippingRateID:   &sharedRateID,
			},
		}}, nil)

	// Called twice (once per item), but same rate id
	tc.mockCart.EXPECT().
		GetShippingRateByID(gomock.Any(), sharedRateID).
		Return(&cartEntities.CartShippingRate{Id: sharedRateID, Amount: decimal.NewFromInt(5), CarrierName: "Carrier", CarrierCode: "CARR"}, nil).Times(2)

	tc.mockCustomer.EXPECT().
		GetCustomer(gomock.Any()).
		Return(&customerEntities.Customer{ID: customerID, StripeID: nullable.StringPtr(customerStripeID)}, nil)

	tc.mockPayment.EXPECT().
		GetProvider().
		Return(providers.ProviderStripe).Times(2)

	tc.mockPayment.EXPECT().
		CreatePayment(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, req stripeEntities.CreatePaymentIntentRequest) {
			assert.Equal(t, expectedTotal, req.Amount)
		}).
		Return(providers.PaymentProviderResponse{ID: "pi_dup", Status: providers.PaymentStatusPending}, nil)

	tc.mockRepo.EXPECT().
		OrderReferenceExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	tc.mockRepo.EXPECT().
		CreateOrder(gomock.Any(), cartID, gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ uuid.UUID, order *entities.Order, _ []*entities.OrderItem) {
			assert.Equal(t, decimal.NewFromInt(5), *order.ShippingRate)
			assert.Equal(t, expectedTotal, order.Total)
			// No order-level carrier fields when request ShippingRateID is not provided
			assert.Empty(t, order.ShippingCarrierName)
			assert.Empty(t, order.ShippingCarrierCode)
		}).
		Return(nil)

	notifyDone := make(chan struct{})
	tc.mockWebhook.EXPECT().
		NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ *webhookEntities.NotifyOrderStatusChangeRequest) { close(notifyDone) }).
		Return(nil)

	salesforceCallDone := make(chan struct{})
	tc.mockInventory.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ inventoryEntities.CreateInventoryOrderRequest) { close(salesforceCallDone) }).
		Return(nil, nil)

	req := &entities.CreateOrderRequest{Body: &entities.CreateOrderRequestBody{AddressID: addressID, StripePaymentMethodID: paymentMethodID}}
	resp, err := s.CreateOrder(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	<-notifyDone
	<-salesforceCallDone
}

func TestCreateOrder_NoShippingRates_TotalNoShipping(t *testing.T) {
	tc := setupTestController(t)
	s := newServiceUnderTest(tc)

	customerID := uuid.New()
	addressID := uuid.New()
	cartID := uuid.New()
	paymentMethodID := "pm_noship"
	customerStripeID := "cus_noship"

	// subtotal: (25*2) + (15*1) = 65
	// tax: 10
	// shipping: 0 (no shipping rates)
	expectedTotal := decimal.NewFromInt(75)

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

	tc.mockAddress.EXPECT().
		GetAddress(gomock.Any(), &addressEntities.GetAddressRequest{AddressID: addressID}).
		Return(&addressEntities.Address{
			FullName:    "John Doe",
			Address:     "123 Main St",
			City:        nullable.StringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: nullable.StringPtr("1234567890"),
		}, nil)

	tc.mockCart.EXPECT().
		GetCart(gomock.Any()).
		Return(&cartEntities.Cart{Id: cartID, TaxAmount: decimal.NewFromInt(10), TaxCurrency: "USD"}, nil)

	tc.mockCart.EXPECT().
		GetCartItems(gomock.Any()).
		Return(&cartEntities.GetCartItemsResponse{Items: []cartEntities.CartItemDetail{
			{
				ProductID:        uuid.New(),
				ProductVariantID: uuid.New(),
				SKU:              "SKU-1",
				Name:             "One",
				Quantity:         2,
				Price:            decimal.NewFromInt(25),
			},
			{
				ProductID:        uuid.New(),
				ProductVariantID: uuid.New(),
				SKU:              "SKU-2",
				Name:             "Two",
				Quantity:         1,
				Price:            decimal.NewFromInt(15),
			},
		}}, nil)

	// No GetShippingRateByID expectations because there are no shipping rates

	tc.mockCustomer.EXPECT().
		GetCustomer(gomock.Any()).
		Return(&customerEntities.Customer{ID: customerID, StripeID: nullable.StringPtr(customerStripeID)}, nil)

	tc.mockPayment.EXPECT().
		GetProvider().
		Return(providers.ProviderStripe).Times(2)

	tc.mockPayment.EXPECT().
		CreatePayment(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, req stripeEntities.CreatePaymentIntentRequest) {
			assert.Equal(t, expectedTotal, req.Amount)
		}).
		Return(providers.PaymentProviderResponse{ID: "pi_noship", Status: providers.PaymentStatusPending}, nil)

	tc.mockRepo.EXPECT().
		OrderReferenceExists(gomock.Any(), gomock.Any()).
		Return(false, nil)

	tc.mockRepo.EXPECT().
		CreateOrder(gomock.Any(), cartID, gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ uuid.UUID, order *entities.Order, items []*entities.OrderItem) {
			assert.Equal(t, expectedTotal, order.Total)
			// Order-level shipping not set when no shipping
			assert.Equal(t, decimal.Zero, *order.ShippingRate)
			// Items should have no shipping fields
			for _, it := range items {
				assert.Empty(t, it.ShippingRateID)
				assert.Empty(t, it.ShippingRate)
				assert.Empty(t, it.ShippingCarrierName)
				assert.Empty(t, it.ShippingCarrierCode)
				assert.Empty(t, it.ShippingServiceType)
				assert.Empty(t, it.ShippingServiceCode)
				assert.Empty(t, it.EstimatedDeliveryDate)
				assert.Empty(t, it.BusinessDaysInTransit)
			}
		}).
		Return(nil)

	notifyDone := make(chan struct{})
	tc.mockWebhook.EXPECT().
		NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ *webhookEntities.NotifyOrderStatusChangeRequest) { close(notifyDone) }).
		Return(nil)

	salesforceCallDone := make(chan struct{})
	tc.mockInventory.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, _ inventoryEntities.CreateInventoryOrderRequest) { close(salesforceCallDone) }).
		Return(nil, nil)

	req := &entities.CreateOrderRequest{Body: &entities.CreateOrderRequestBody{AddressID: addressID, StripePaymentMethodID: paymentMethodID}}
	resp, err := s.CreateOrder(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	<-notifyDone
	<-salesforceCallDone
}

func TestProcessPaymentSucceeded_WithStripe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		customerID := uuid.New()
		paymentID := "pi_123"
		orderID := uuid.New()
		salesforceID := "123456"
		productID := uuid.New()
		productVariantID := uuid.New()

		ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			GetOrderByStripePaymentIntentID(gomock.Any(), paymentID).
			Return(&entities.Order{
				ID:                    orderID,
				CustomerID:            customerID,
				StripePaymentIntentID: nullable.StringPtr(paymentID),
				Status:                entities.Pending,
				SalesforceID:          salesforceID,
			}, nil)

		tc.mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, updates map[string]interface{}, orderID string, customerID string) {
				assert.Equal(t, orderID, orderID)
				assert.Equal(t, customerID, customerID)
				assert.Equal(t, entities.PaymentSuccess, updates["status"])
			}).
			Return(nil)

		tc.mockCustomer.EXPECT().
			GetCustomerByID(gomock.Any(), customerID.String()).
			Return(&customerEntities.Customer{
				ID:           customerID,
				SalesforceID: nullable.StringPtr(salesforceID),
			}, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return([]*entities.OrderItem{
				{
					ProductID: productID,
				},
			}, nil)

		notifyCallDone := make(chan struct{})
		tc.mockWebhook.EXPECT().
			NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *webhookEntities.NotifyOrderStatusChangeRequest) {
				assert.Equal(t, customerID.String(), req.CustomerID)
				assert.Equal(t, entities.PaymentSuccess.String(), req.Status)
				close(notifyCallDone)
			}).
			Return(nil)

		salesforceCallDone := make(chan struct{})
		tc.mockInventory.EXPECT().
			UpdateOrderStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req inventoryEntities.UpdateInventoryOrderStatusRequest) {
				assert.Equal(t, salesforceID, req.Order.SalesforceID)
				assert.Equal(t, salesforceID, *req.Customer.SalesforceID)
				assert.Equal(t, entities.PaymentSuccess.String(), req.Status)
				close(salesforceCallDone)
			}).
			Return(nil)

		tc.mockProduct.EXPECT().
			GetProductVariantByID(gomock.Any(), gomock.Any()).
			Return(&productEntities.ProductVariant{
				ID:        productVariantID,
				ProductID: productID,
			}, nil)

		tc.mockWishlist.EXPECT().
			BulkRemoveFromWishlist(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *wishlistEntities.BulkRemoveFromWishlistRequest) {
				assert.Equal(t, customerID, req.CustomerID)
				assert.Equal(t, []uuid.UUID{productID}, req.ProductIDs)
			}).
			Return(nil)

		err := s.ProcessPaymentSucceeded(ctx, paymentID)

		assert.NoError(t, err)
		// wait for async calls to be done
		<-notifyCallDone
		<-salesforceCallDone
	})

	t.Run("error to get order by payment id", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		customerID := uuid.New()
		paymentID := "12346"

		ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			GetOrderByStripePaymentIntentID(gomock.Any(), paymentID).
			Return(nil, errors.New("order not found"))

		err := s.ProcessPaymentSucceeded(ctx, paymentID)

		assert.ErrorContains(t, err, moduleErrors.NewAPIError("ORDER_NOT_FOUND_BY_PAYMENT_ID").Error())
	})
}

func TestProcessPaymentSucceeded_WithAuthorizeNet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		customerID := uuid.New()
		paymentID := "123456"
		orderID := uuid.New()
		salesforceID := "12346"
		productID := uuid.New()
		productVariantID := uuid.New()

		ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderAuthorizeNet)

		tc.mockRepo.EXPECT().
			GetOrderByAuthorizeNetPaymentID(gomock.Any(), paymentID).
			Return(&entities.Order{
				ID:                    orderID,
				CustomerID:            customerID,
				AuthorizeNetPaymentID: nullable.StringPtr(paymentID),
				Status:                entities.Pending,
				SalesforceID:          salesforceID,
			}, nil)

		tc.mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, updates map[string]interface{}, orderID string, customerID string) {
				assert.Equal(t, orderID, orderID)
				assert.Equal(t, customerID, customerID)
				assert.Equal(t, entities.PaymentSuccess, updates["status"])
			}).
			Return(nil)

		tc.mockCustomer.EXPECT().
			GetCustomerByID(gomock.Any(), customerID.String()).
			Return(&customerEntities.Customer{
				ID:           customerID,
				SalesforceID: nullable.StringPtr(salesforceID),
			}, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return([]*entities.OrderItem{
				{
					ProductID: productID,
				},
			}, nil)

		notifyCallDone := make(chan struct{})
		tc.mockWebhook.EXPECT().
			NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *webhookEntities.NotifyOrderStatusChangeRequest) {
				assert.Equal(t, customerID.String(), req.CustomerID)
				assert.Equal(t, entities.PaymentSuccess.String(), req.Status)
				close(notifyCallDone)
			}).
			Return(nil)

		salesforceCallDone := make(chan struct{})
		tc.mockInventory.EXPECT().
			UpdateOrderStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req inventoryEntities.UpdateInventoryOrderStatusRequest) {
				assert.Equal(t, salesforceID, req.Order.SalesforceID)
				assert.Equal(t, salesforceID, *req.Customer.SalesforceID)
				assert.Equal(t, entities.PaymentSuccess.String(), req.Status)
				close(salesforceCallDone)
			}).
			Return(nil)

		tc.mockProduct.EXPECT().
			GetProductVariantByID(gomock.Any(), gomock.Any()).
			Return(&productEntities.ProductVariant{
				ID:        productVariantID,
				ProductID: productID,
			}, nil)

		tc.mockWishlist.EXPECT().
			BulkRemoveFromWishlist(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *wishlistEntities.BulkRemoveFromWishlistRequest) {
				assert.Equal(t, customerID, req.CustomerID)
				assert.Equal(t, []uuid.UUID{productID}, req.ProductIDs)
			}).
			Return(nil)

		err := s.ProcessPaymentSucceeded(ctx, paymentID)

		assert.NoError(t, err)
		// wait for async calls to be done
		<-notifyCallDone
		<-salesforceCallDone
	})

	t.Run("error to get order by payment id", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		customerID := uuid.New()
		paymentID := "12346"

		ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderAuthorizeNet)

		tc.mockRepo.EXPECT().
			GetOrderByAuthorizeNetPaymentID(gomock.Any(), paymentID).
			Return(nil, errors.New("order not found"))

		err := s.ProcessPaymentSucceeded(ctx, paymentID)

		assert.ErrorContains(t, err, moduleErrors.NewAPIError("ORDER_NOT_FOUND_BY_PAYMENT_ID").Error())
	})
}

func TestUpdateOrder(t *testing.T) {
	t.Run("success with status change", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		newStatus := "shipped"
		trackingNumber := "TRK123456"
		trackingURL := "https://tracking.example.com/TRK123456"

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
		}

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any(), orderID.String(), customerID.String()).
			Do(func(_ context.Context, data map[string]interface{}, orderID, customerID string) {
				status, ok := data["status"].(*string)
				assert.True(t, ok)
				assert.Equal(t, newStatus, *status)
			}).
			Return(nil)

		notifyCallDone := make(chan struct{})
		tc.mockWebhook.EXPECT().
			NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *webhookEntities.NotifyOrderStatusChangeRequest) {
				assert.Equal(t, customerID.String(), req.CustomerID)
				assert.Equal(t, orderID.String(), req.OrderID)
				assert.Equal(t, orderRef, req.OrderReference)
				assert.Equal(t, newStatus, req.Status)
				close(notifyCallDone)
			}).
			Return(nil)

		req := &entities.UpdateOrderRequest{
			OrderReference: orderRef,
			Body: &entities.UpdateOrderRequestBody{
				Status:                    &newStatus,
				FulfillmentTrackingNumber: &trackingNumber,
				FulfillmentTrackingURL:    &trackingURL,
			},
		}

		err := s.UpdateOrder(ctx, req)

		assert.NoError(t, err)
		// wait for async notify call to be done
		<-notifyCallDone
	})

	t.Run("success without status change", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		trackingNumber := "TRK123456"
		shipmentDate := time.Now()

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
		}

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any(), orderID.String(), customerID.String()).
			Do(func(_ context.Context, data map[string]interface{}, orderID, customerID string) {
				// status should not be in the update data
				_, hasStatus := data["status"]
				assert.False(t, hasStatus)
			}).
			Return(nil)

		// No webhook notification should be called since status didn't change
		tc.mockWebhook.EXPECT().
			NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
			Times(0)

		req := &entities.UpdateOrderRequest{
			OrderReference: orderRef,
			Body: &entities.UpdateOrderRequestBody{
				FulfillmentTrackingNumber: &trackingNumber,
				FulfillmentShipmentDate:   &shipmentDate,
			},
		}

		err := s.UpdateOrder(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("success with order items update", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		itemID := uuid.New()
		itemStatus := entities.OrderItemStatus("shipped")

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
		}

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any(), orderID.String(), customerID.String()).
			Do(func(_ context.Context, data map[string]interface{}, orderID, customerID string) {
				items, ok := data["items"].([]map[string]interface{})
				assert.True(t, ok)
				assert.Len(t, items, 1)
				assert.Equal(t, itemID.String(), items[0]["id"])

				status, ok := items[0]["status"].(*entities.OrderItemStatus)
				assert.True(t, ok)
				assert.Equal(t, itemStatus, *status)
			}).
			Return(nil)

		req := &entities.UpdateOrderRequest{
			OrderReference: orderRef,
			Body: &entities.UpdateOrderRequestBody{
				Items: []*entities.Item{
					{
						ID:     itemID.String(),
						Status: &itemStatus,
					},
				},
			},
		}

		err := s.UpdateOrder(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("error getting order by reference", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderRef := "ORD123456"
		ctx := context.Background()

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(nil, errors.New("order not found"))

		req := &entities.UpdateOrderRequest{
			OrderReference: orderRef,
			Body: &entities.UpdateOrderRequestBody{
				Status: nullable.StringPtr("shipped"),
			},
		}

		err := s.UpdateOrder(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order not found")
	})

	t.Run("error updating order", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		newStatus := "shipped"

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
		}

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any(), orderID.String(), customerID.String()).
			Return(errors.New("database error"))

		req := &entities.UpdateOrderRequest{
			OrderReference: orderRef,
			Body: &entities.UpdateOrderRequestBody{
				Status: &newStatus,
			},
		}

		err := s.UpdateOrder(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("success with same status - no notification", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		currentStatus := entities.PaymentSuccess.String()

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
		}

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any(), orderID.String(), customerID.String()).
			Return(nil)

		// No webhook notification should be called since status is the same
		tc.mockWebhook.EXPECT().
			NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
			Times(0)

		req := &entities.UpdateOrderRequest{
			OrderReference: orderRef,
			Body: &entities.UpdateOrderRequestBody{
				Status: &currentStatus,
			},
		}

		err := s.UpdateOrder(ctx, req)

		assert.NoError(t, err)
	})
}

func TestRefundOrder_WithStripe(t *testing.T) {
	t.Run("success with stripe refund", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		orderItemID := uuid.New()
		paymentIntentID := "pi_123"
		refundID := "re_123"
		itemSKU := "SKU123"
		itemPrice := decimal.NewFromInt(50)
		refundQuantity := 1
		expectedRefundAmount := itemPrice.Mul(decimal.NewFromInt(int64(refundQuantity)))

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:                    orderID,
			CustomerID:            customerID,
			OrderReference:        orderRef,
			Status:                entities.PaymentSuccess,
			Total:                 decimal.NewFromInt(100),
			StripePaymentIntentID: &paymentIntentID,
		}

		orderItems := []*entities.OrderItem{
			{
				ID:       orderItemID,
				OrderID:  orderID,
				SKU:      itemSKU,
				Price:    itemPrice,
				Quantity: 2,
				Status:   entities.ItemDelivered,
			},
		}

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(orderItems, nil)

		tc.mockPayment.EXPECT().
			Refund(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *stripeEntities.RefundRequest) {
				assert.Equal(t, paymentIntentID, req.PaymentIntentId)
				assert.Equal(t, expectedRefundAmount, req.Amount)
			}).
			Return(&providers.RefundResponse{
				ID:     refundID,
				Status: stripeEntities.StripeRefundSucceeded,
			}, nil)

		tc.mockRepo.EXPECT().
			UpdateOrderWithOrderItems(gomock.Any(), orderID, gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ uuid.UUID, orderData map[string]interface{}, itemsData map[string]interface{}) {
				// Order should not be marked as fully refunded since we're only refunding 1 out of 2 items
				_, hasStatus := orderData["status"]
				assert.False(t, hasStatus)

				// Check item refund data
				itemData, exists := itemsData[orderItemID.String()]
				assert.True(t, exists)

				itemMap := itemData.(map[string]interface{})
				assert.Equal(t, entities.ItemInitiatedRefund.String(), itemMap["status"])
				assert.Equal(t, refundID, itemMap["stripe_refund_id"])
				assert.Equal(t, expectedRefundAmount.InexactFloat64(), itemMap["stripe_refund_amount"])
			}).
			Return(nil)

		notifyCallDone := make(chan struct{})
		tc.mockWebhook.EXPECT().
			NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *webhookEntities.NotifyOrderStatusChangeRequest) {
				assert.Equal(t, customerID.String(), req.CustomerID)
				assert.Equal(t, orderID.String(), req.OrderID)
				assert.Equal(t, orderRef, req.OrderReference)
				assert.Equal(t, entities.Refunded.String(), req.Status)
				close(notifyCallDone)
			}).
			Return(nil)

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      itemSKU,
						Quantity: refundQuantity,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedRefundAmount, resp.TotalRefundableAmount)
		assert.Len(t, resp.RefundableItems, 1)
		assert.Equal(t, orderItemID.String(), resp.RefundableItems[0].ItemId)
		assert.Equal(t, itemSKU, resp.RefundableItems[0].Sku)
		assert.Equal(t, refundQuantity, resp.RefundableItems[0].Quantity)
		assert.True(t, resp.RefundableItems[0].RefundInitiated)
		// wait for async notify call to be done
		<-notifyCallDone
	})

	t.Run("success with full order refund", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		orderItemID := uuid.New()
		paymentIntentID := "pi_123"
		refundID := "re_123"
		itemSKU := "SKU123"
		itemPrice := decimal.NewFromInt(50)
		refundQuantity := 2
		expectedRefundAmount := itemPrice.Mul(decimal.NewFromInt(int64(refundQuantity)))

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:                    orderID,
			CustomerID:            customerID,
			OrderReference:        orderRef,
			Status:                entities.PaymentSuccess,
			Total:                 decimal.NewFromInt(100),
			StripePaymentIntentID: &paymentIntentID,
		}

		orderItems := []*entities.OrderItem{
			{
				ID:       orderItemID,
				OrderID:  orderID,
				SKU:      itemSKU,
				Price:    itemPrice,
				Quantity: 2,
				Status:   entities.ItemDelivered,
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(orderItems, nil)

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockPayment.EXPECT().
			Refund(gomock.Any(), gomock.Any()).
			Return(&providers.RefundResponse{
				ID:     refundID,
				Status: stripeEntities.StripeRefundSucceeded,
			}, nil)

		tc.mockRepo.EXPECT().
			UpdateOrderWithOrderItems(gomock.Any(), orderID, gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ uuid.UUID, orderData map[string]interface{}, itemsData map[string]interface{}) {
				assert.Equal(t, expectedRefundAmount, orderData["stripe_refund_total"])
			}).
			Return(nil)

		notifyCallDone := make(chan struct{})
		tc.mockWebhook.EXPECT().
			NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *webhookEntities.NotifyOrderStatusChangeRequest) {
				assert.Equal(t, entities.Refunded.String(), req.Status)
				close(notifyCallDone)
			}).
			Return(nil)

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      itemSKU,
						Quantity: refundQuantity,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedRefundAmount, resp.TotalRefundableAmount)
		// wait for async notify call to be done
		<-notifyCallDone
	})

	t.Run("error order not found", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderRef := "ORD123456"
		ctx := context.Background()

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(nil, errors.New("order not found"))

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      "SKU123",
						Quantity: 1,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), moduleErrors.NewAPIError("ORDER_NOT_FOUND").Error())
	})

	t.Run("error order already refunded", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.Refunded,
		}

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      "SKU123",
						Quantity: 1,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "Order is not eligible for refund")
	})

	t.Run("error no refundable items found", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
			Total:          decimal.NewFromInt(100),
		}

		orderItems := []*entities.OrderItem{
			{
				ID:       uuid.New(),
				OrderID:  orderID,
				SKU:      "DIFFERENT_SKU",
				Price:    decimal.NewFromInt(50),
				Quantity: 2,
				Status:   entities.ItemDelivered,
			},
		}

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(orderItems, nil)

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      "SKU123",
						Quantity: 1,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "No refundable items found or amount is zero")
	})

	t.Run("error refundable amount exceeds order total", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		orderItemID := uuid.New()
		itemSKU := "SKU123"
		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
			Total:          decimal.NewFromInt(50), // Lower than refund amount
		}

		orderItems := []*entities.OrderItem{
			{
				ID:       orderItemID,
				OrderID:  orderID,
				SKU:      itemSKU,
				Price:    decimal.NewFromInt(100), // Higher than order total
				Quantity: 2,
				Status:   entities.ItemDelivered,
			},
		}

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(orderItems, nil)

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      itemSKU,
						Quantity: 1,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "Refundable amount exceeds order total")
	})

	t.Run("error stripe refund failed", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		orderItemID := uuid.New()
		paymentIntentID := "pi_123"
		itemSKU := "SKU123"
		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:                    orderID,
			CustomerID:            customerID,
			OrderReference:        orderRef,
			Status:                entities.PaymentSuccess,
			Total:                 decimal.NewFromInt(100),
			StripePaymentIntentID: &paymentIntentID,
		}

		orderItems := []*entities.OrderItem{
			{
				ID:       orderItemID,
				OrderID:  orderID,
				SKU:      itemSKU,
				Price:    decimal.NewFromInt(50),
				Quantity: 2,
				Status:   entities.ItemDelivered,
			},
		}

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(orderItems, nil)

		tc.mockPayment.EXPECT().
			Refund(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("stripe refund failed"))

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      itemSKU,
						Quantity: 1,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "Error processing refund via Stripe")
	})

	t.Run("error unsupported payment provider", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		orderItemID := uuid.New()
		itemSKU := "SKU123"
		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
			Total:          decimal.NewFromInt(100),
		}

		orderItems := []*entities.OrderItem{
			{
				ID:       orderItemID,
				OrderID:  orderID,
				SKU:      itemSKU,
				Price:    decimal.NewFromInt(50),
				Quantity: 2,
				Status:   entities.ItemDelivered,
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(orderItems, nil)

		randomProvider := providers.ProviderType("random_provider")
		tc.mockPayment.EXPECT().
			GetProvider().
			Return(randomProvider)

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      itemSKU,
						Quantity: 1,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "Payment provider not supported for refunds")
	})

	t.Run("skip already refunded items", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		orderItemID1 := uuid.New()
		orderItemID2 := uuid.New()
		itemSKU := "SKU123"
		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
			Total:          decimal.NewFromInt(100),
		}

		orderItems := []*entities.OrderItem{
			{
				ID:       orderItemID1,
				OrderID:  orderID,
				SKU:      itemSKU,
				Price:    decimal.NewFromInt(50),
				Quantity: 1,
				Status:   entities.ItemRefunded, // Already refunded
			},
			{
				ID:       orderItemID2,
				OrderID:  orderID,
				SKU:      itemSKU,
				Price:    decimal.NewFromInt(50),
				Quantity: 1,
				Status:   entities.ItemDelivered, // Available for refund
			},
		}

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(orderItems, nil)

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      itemSKU,
						Quantity: 2, // Requesting more than available
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "No refundable items found or amount is zero")
	})

	t.Run("partial refund does not change order status", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		orderRef := "ORD123456"
		orderItemID1 := uuid.New()
		orderItemID2 := uuid.New()
		paymentIntentID := "pi_123"
		refundID := "re_123"
		itemSKU1 := "SKU123"
		itemSKU2 := "SKU456"
		itemPrice := decimal.NewFromInt(50)
		refundQuantity := 1
		expectedRefundAmount := itemPrice.Mul(decimal.NewFromInt(int64(refundQuantity)))

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:                    orderID,
			CustomerID:            customerID,
			OrderReference:        orderRef,
			Status:                entities.PaymentSuccess,
			Total:                 decimal.NewFromInt(150),
			StripePaymentIntentID: &paymentIntentID,
		}

		// Total items: 3 (item1: 1, item2: 2), refunding only 1 item
		orderItems := []*entities.OrderItem{
			{
				ID:       orderItemID1,
				OrderID:  orderID,
				SKU:      itemSKU1,
				Price:    itemPrice,
				Quantity: 1,
				Status:   entities.ItemDelivered,
			},
			{
				ID:       orderItemID2,
				OrderID:  orderID,
				SKU:      itemSKU2,
				Price:    itemPrice,
				Quantity: 2,
				Status:   entities.ItemDelivered,
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderByReference(gomock.Any(), orderRef).
			Return(existingOrder, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(orderItems, nil)

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockPayment.EXPECT().
			Refund(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *stripeEntities.RefundRequest) {
				assert.Equal(t, paymentIntentID, req.PaymentIntentId)
				assert.Equal(t, expectedRefundAmount, req.Amount)
			}).
			Return(&providers.RefundResponse{
				ID:     refundID,
				Status: stripeEntities.StripeRefundSucceeded,
			}, nil)

		tc.mockRepo.EXPECT().
			UpdateOrderWithOrderItems(gomock.Any(), orderID, gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ uuid.UUID, orderData map[string]interface{}, itemsData map[string]interface{}) {
				// Order should NOT be marked as fully refunded since we're only refunding 1 out of 3 total items
				_, hasStatus := orderData["status"]
				assert.False(t, hasStatus, "Order status should not change for partial refund")

				// Check item refund data
				itemData, exists := itemsData[orderItemID1.String()]
				assert.True(t, exists)

				itemMap := itemData.(map[string]interface{})
				assert.Equal(t, entities.ItemInitiatedRefund.String(), itemMap["status"])
				assert.Equal(t, refundID, itemMap["stripe_refund_id"])
				assert.Equal(t, expectedRefundAmount.InexactFloat64(), itemMap["stripe_refund_amount"])
			}).
			Return(nil)

		notifyCallDone := make(chan struct{})
		tc.mockWebhook.EXPECT().
			NotifyOrderStatusChange(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *webhookEntities.NotifyOrderStatusChangeRequest) {
				assert.Equal(t, customerID.String(), req.CustomerID)
				assert.Equal(t, orderID.String(), req.OrderID)
				assert.Equal(t, orderRef, req.OrderReference)
				assert.Equal(t, entities.Refunded.String(), req.Status)
				close(notifyCallDone)
			}).
			Return(nil)

		req := &entities.RefundOrderRequest{
			OrderReference: orderRef,
			Body: &entities.RefundOrderRequestBody{
				Items: []*entities.RefundItem{
					{
						Sku:      itemSKU1,
						Quantity: refundQuantity,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedRefundAmount, resp.TotalRefundableAmount)
		assert.Len(t, resp.RefundableItems, 1)
		assert.Equal(t, orderItemID1.String(), resp.RefundableItems[0].ItemId)
		assert.Equal(t, itemSKU1, resp.RefundableItems[0].Sku)
		assert.Equal(t, refundQuantity, resp.RefundableItems[0].Quantity)
		assert.True(t, resp.RefundableItems[0].RefundInitiated)
		// wait for async notify call to be done
		<-notifyCallDone
	})

}

func TestProcessRefundSucceeded(t *testing.T) {
	t.Run("success with stripe refund - partial refund", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		refundID := "re_123"
		refundAmount := decimal.NewFromInt(50)
		orderItemID1 := uuid.New()
		orderItemID2 := uuid.New()

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:         orderID,
			CustomerID: customerID,
			Status:     entities.PaymentSuccess,
		}

		orderItems := []*entities.OrderItem{
			{
				ID:             orderItemID1,
				OrderID:        orderID,
				Status:         entities.ItemInitiatedRefund,
				StripeRefundID: refundID,
			},
			{
				ID:             orderItemID2,
				OrderID:        orderID,
				Status:         entities.ItemDelivered,
				StripeRefundID: "",
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderItemsByStripeRefundID(gomock.Any(), refundID).
			Return(orderItems, nil)

		tc.mockRepo.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(existingOrder, nil)

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			UpdateOrderWithOrderItems(gomock.Any(), orderID, gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ uuid.UUID, orderData map[string]interface{}, itemsData map[string]interface{}) {
				// Order should NOT be marked as refunded since not all items are refunded
				_, hasStatus := orderData["status"]
				assert.False(t, hasStatus, "Order status should not change for partial refund")

				// Check item refund data
				itemData, exists := itemsData[orderItemID1.String()]
				assert.True(t, exists)

				itemMap := itemData.(map[string]interface{})
				assert.Equal(t, entities.ItemRefunded, itemMap["status"])
				assert.Equal(t, refundAmount.InexactFloat64(), itemMap["stripe_refund_amount"])
			}).
			Return(nil)

		err := s.ProcessRefundSucceeded(ctx, refundID, refundAmount)

		assert.NoError(t, err)
	})

	t.Run("success with stripe refund - full refund", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		refundID := "re_123"
		refundAmount := decimal.NewFromInt(100)
		orderItemID1 := uuid.New()
		orderItemID2 := uuid.New()
		existingRefundTotal := decimal.NewFromInt(50)

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:                orderID,
			CustomerID:        customerID,
			Status:            entities.PaymentSuccess,
			StripeRefundTotal: &existingRefundTotal,
		}

		orderItems := []*entities.OrderItem{
			{
				ID:             orderItemID1,
				OrderID:        orderID,
				Status:         entities.ItemInitiatedRefund,
				StripeRefundID: refundID,
			},
			{
				ID:             orderItemID2,
				OrderID:        orderID,
				Status:         entities.ItemRefunded, // Already refunded
				StripeRefundID: "re_456",
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderItemsByStripeRefundID(gomock.Any(), refundID).
			Return(orderItems, nil)

		tc.mockRepo.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(existingOrder, nil)

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			UpdateOrderWithOrderItems(gomock.Any(), orderID, gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ uuid.UUID, orderData map[string]interface{}, itemsData map[string]interface{}) {
				// Order SHOULD be marked as refunded since all items are now refunded
				status, hasStatus := orderData["status"]
				assert.True(t, hasStatus, "Order status should change for full refund")
				assert.Equal(t, entities.Refunded, status)

				// Check updated refund total
				expectedTotal := existingRefundTotal.Add(refundAmount)
				assert.Equal(t, expectedTotal, orderData["stripe_refund_total"])

				// Check item refund data
				itemData, exists := itemsData[orderItemID1.String()]
				assert.True(t, exists)

				itemMap := itemData.(map[string]interface{})
				assert.Equal(t, entities.ItemRefunded, itemMap["status"])
				assert.Equal(t, refundAmount.InexactFloat64(), itemMap["stripe_refund_amount"])
			}).
			Return(nil)

		err := s.ProcessRefundSucceeded(ctx, refundID, refundAmount)

		assert.NoError(t, err)
	})

	t.Run("error - no order items found for refund ID", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		refundID := "re_123"
		refundAmount := decimal.NewFromInt(50)

		ctx := context.Background()

		tc.mockRepo.EXPECT().
			GetOrderItemsByStripeRefundID(gomock.Any(), refundID).
			Return([]*entities.OrderItem{}, nil)

		err := s.ProcessRefundSucceeded(ctx, refundID, refundAmount)

		assert.NoError(t, err) // Should not return error but log and return nil
	})

	t.Run("error - failed to fetch order items by refund ID", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		refundID := "re_123"
		refundAmount := decimal.NewFromInt(50)

		ctx := context.Background()

		tc.mockRepo.EXPECT().
			GetOrderItemsByStripeRefundID(gomock.Any(), refundID).
			Return(nil, errors.New("database error"))

		err := s.ProcessRefundSucceeded(ctx, refundID, refundAmount)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), moduleErrors.NewAPIError("ORDER_ITEMS_NOT_FOUND_BY_REFUND_ID").Error())
	})

	t.Run("error - failed to fetch order by ID", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		refundID := "re_123"
		refundAmount := decimal.NewFromInt(50)
		orderItemID := uuid.New()

		ctx := context.Background()

		orderItems := []*entities.OrderItem{
			{
				ID:             orderItemID,
				OrderID:        orderID,
				Status:         entities.ItemInitiatedRefund,
				StripeRefundID: refundID,
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderItemsByStripeRefundID(gomock.Any(), refundID).
			Return(orderItems, nil)

		tc.mockRepo.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(nil, errors.New("order not found"))

		err := s.ProcessRefundSucceeded(ctx, refundID, refundAmount)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), moduleErrors.NewAPIError("ORDER_NOT_FOUND_BY_ID").Error())
	})

	t.Run("order already in refunded state", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		refundID := "re_123"
		refundAmount := decimal.NewFromInt(50)
		orderItemID := uuid.New()

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:         orderID,
			CustomerID: customerID,
			Status:     entities.Refunded, // Already refunded
		}

		orderItems := []*entities.OrderItem{
			{
				ID:             orderItemID,
				OrderID:        orderID,
				Status:         entities.ItemInitiatedRefund,
				StripeRefundID: refundID,
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderItemsByStripeRefundID(gomock.Any(), refundID).
			Return(orderItems, nil)

		tc.mockRepo.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(existingOrder, nil)

		err := s.ProcessRefundSucceeded(ctx, refundID, refundAmount)

		assert.NoError(t, err) // Should not return error but log and return nil
	})

	t.Run("success with new refund total when no existing refund", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		refundID := "re_123"
		refundAmount := decimal.NewFromInt(100)
		orderItemID := uuid.New()

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:                orderID,
			CustomerID:        customerID,
			Status:            entities.PaymentSuccess,
			StripeRefundTotal: nil, // No existing refund
		}

		orderItems := []*entities.OrderItem{
			{
				ID:             orderItemID,
				OrderID:        orderID,
				Status:         entities.ItemInitiatedRefund,
				StripeRefundID: refundID,
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderItemsByStripeRefundID(gomock.Any(), refundID).
			Return(orderItems, nil)

		tc.mockRepo.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(existingOrder, nil)

		tc.mockPayment.EXPECT().
			GetProvider().
			Return(providers.ProviderStripe)

		tc.mockRepo.EXPECT().
			UpdateOrderWithOrderItems(gomock.Any(), orderID, gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ uuid.UUID, orderData map[string]interface{}, itemsData map[string]interface{}) {
				// Order should be marked as refunded since all items are now refunded
				status, hasStatus := orderData["status"]
				assert.True(t, hasStatus)
				assert.Equal(t, entities.Refunded, status)

				// Check new refund total
				assert.Equal(t, refundAmount, orderData["stripe_refund_total"])

				// Check item refund data
				itemData, exists := itemsData[orderItemID.String()]
				assert.True(t, exists)

				itemMap := itemData.(map[string]interface{})
				assert.Equal(t, entities.ItemRefunded, itemMap["status"])
				assert.Equal(t, refundAmount.InexactFloat64(), itemMap["stripe_refund_amount"])
			}).
			Return(nil)

		err := s.ProcessRefundSucceeded(ctx, refundID, refundAmount)

		assert.NoError(t, err)
	})

	t.Run("unsupported payment provider", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		customerID := uuid.New()
		refundID := "re_123"
		refundAmount := decimal.NewFromInt(50)
		orderItemID := uuid.New()

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:         orderID,
			CustomerID: customerID,
			Status:     entities.PaymentSuccess,
		}

		orderItems := []*entities.OrderItem{
			{
				ID:             orderItemID,
				OrderID:        orderID,
				Status:         entities.ItemInitiatedRefund,
				StripeRefundID: refundID,
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderItemsByStripeRefundID(gomock.Any(), refundID).
			Return(orderItems, nil)

		tc.mockRepo.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(existingOrder, nil)

		randomProvider := providers.ProviderType("random_provider")
		tc.mockPayment.EXPECT().
			GetProvider().
			Return(randomProvider)

		err := s.ProcessRefundSucceeded(ctx, refundID, refundAmount)

		assert.NoError(t, err) // Should not return error but log the unsupported provider
	})
}

func TestListOrders(t *testing.T) {
	t.Run("success returns orders for customer", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		customerID := uuid.New()
		ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

		req := &entities.ListOrdersRequest{
			Limit: 20,
		}

		o1 := &entities.Order{
			ID:             uuid.New(),
			CustomerID:     customerID,
			OrderReference: "ORD-001",
			Status:         entities.Pending,
		}
		o2 := &entities.Order{
			ID:             uuid.New(),
			CustomerID:     customerID,
			OrderReference: "ORD-002",
			Status:         entities.PaymentSuccess,
		}

		// Expect repository to return two orders for the customer
		tc.mockRepo.EXPECT().ListOrders(ctx, customerID, req.Limit, req.Cursor, req.IncludeItems).Return([]*entities.Order{o1, o2}, "", nil)

		resp, err := s.ListOrders(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// Minimal assertions to avoid coupling to response shape beyond orders
		assert.Len(t, resp.Orders, 2)
		assert.Equal(t, "ORD-001", resp.Orders[0].OrderReference)
		assert.Equal(t, "ORD-002", resp.Orders[1].OrderReference)
	})

	t.Run("repository error", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		customerID := uuid.New()
		ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())
		req := &entities.ListOrdersRequest{
			Limit: 20,
		}

		tc.mockRepo.EXPECT().
			ListOrders(ctx, customerID, req.Limit, req.Cursor, req.IncludeItems).
			Return(nil, "", errors.New("database error"))

		resp, err := s.ListOrders(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("missing customer ID in context", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		ctx := context.Background()
		req := &entities.ListOrdersRequest{
			Limit: 20,
		}

		resp, err := s.ListOrders(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetOrder(t *testing.T) {
	t.Run("success - return order with items", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		customerID := uuid.New()
		orderID := uuid.New()
		orderRef := "ORD-123456"
		orderItemID1 := uuid.New()
		orderItemID2 := uuid.New()
		productID1 := uuid.New()
		productID2 := uuid.New()
		productVariantID1 := uuid.New()
		productVariantID2 := uuid.New()

		ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

		tax := decimal.NewFromInt(10)
		shipping := decimal.NewFromInt(5)

		order := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
			Total:          decimal.NewFromInt(100),
			Subtotal:       decimal.NewFromInt(85),
			ShippingRate:   &shipping,
			TaxAmount:      tax,
			CreatedAt:      time.Now(),
		}

		orderItems := []*entities.OrderItem{
			{
				ID:               orderItemID1,
				OrderID:          orderID,
				ProductID:        productID1,
				ProductVariantID: productVariantID1,
				SKU:              "SKU-001",
				Name:             "Product 1",
				Quantity:         1,
				Price:            decimal.NewFromInt(50),
				Status:           entities.ItemDelivered,
			},
			{
				ID:               orderItemID2,
				OrderID:          orderID,
				ProductID:        productID2,
				ProductVariantID: productVariantID2,
				SKU:              "SKU-002",
				Name:             "Product 2",
				Quantity:         1,
				Price:            decimal.NewFromInt(35),
				Status:           entities.ItemDelivered,
			},
		}

		tc.mockRepo.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(order, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(orderItems, nil)

		req := &entities.GetOrderRequest{
			OrderID: orderID,
		}

		resp, err := s.GetOrder(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, orderRef, resp.Order.OrderReference)
		assert.Equal(t, entities.PaymentSuccess, resp.Order.Status)
		assert.Equal(t, decimal.NewFromInt(100), resp.Order.Total)
		assert.Equal(t, decimal.NewFromInt(85), resp.Order.Subtotal)
		assert.Equal(t, decimal.NewFromInt(5), *resp.Order.ShippingRate)
		assert.Equal(t, decimal.NewFromInt(10), resp.Order.TaxAmount)
		assert.Len(t, resp.OrderItems, 2)
	})

	t.Run("error - order not found", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		customerID := uuid.New()
		orderID := uuid.New()

		ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

		tc.mockRepo.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(nil, moduleErrors.NewAPIError("ORDER_NOT_FOUND"))

		req := &entities.GetOrderRequest{
			OrderID: orderID,
		}

		resp, err := s.GetOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), moduleErrors.NewAPIError("ORDER_NOT_FOUND").Error())
	})

	t.Run("error - failed to get order items", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		customerID := uuid.New()
		orderID := uuid.New()
		orderRef := "ORD-123456"

		ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

		order := &entities.Order{
			ID:             orderID,
			CustomerID:     customerID,
			OrderReference: orderRef,
			Status:         entities.PaymentSuccess,
			Total:          decimal.NewFromInt(100),
			CreatedAt:      time.Now(),
		}

		tc.mockRepo.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(order, nil)

		tc.mockRepo.EXPECT().
			GetOrderItemsByID(gomock.Any(), orderID).
			Return(nil, errors.New("database error"))

		req := &entities.GetOrderRequest{
			OrderID: orderID,
		}

		resp, err := s.GetOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), moduleErrors.NewAPIError("ORDER_ERROR_GETTING_ITEMS").Error())
	})

	t.Run("error - missing customer ID in context", func(t *testing.T) {
		tc := setupTestController(t)
		s := newServiceUnderTest(tc)

		orderID := uuid.New()
		ctx := context.Background() // No customer ID in context

		req := &entities.GetOrderRequest{
			OrderID: orderID,
		}

		resp, err := s.GetOrder(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), sharedErrors.NewAPIError("CUSTOMER_ID_REQUIRED").Error())
	})
}
