package service

import (
	"context"
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
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/repository"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	webhookclient "github.com/nurdsoft/nurd-commerce-core/internal/webhook/client"
	webhookEntities "github.com/nurdsoft/nurd-commerce-core/internal/webhook/entities"
	wishlistclient "github.com/nurdsoft/nurd-commerce-core/internal/wishlist/wishlistclient"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	salesforceclient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment"
	authorizenetEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type testController struct {
	mockRepo       *repository.MockRepository
	mockCustomer   *customerclient.MockClient
	mockCart       *cartclient.MockClient
	mockPayment    *payment.MockClient
	mockWishlist   *wishlistclient.MockClient
	mockSalesforce *salesforceclient.MockClient
	mockAddress    *addressclient.MockClient
	mockProduct    *productclient.MockClient
	mockWebhook    *webhookclient.MockClient
}

func setupTestController(t *testing.T) *testController {
	ctrl := gomock.NewController(t)

	return &testController{
		mockRepo:       repository.NewMockRepository(ctrl),
		mockCustomer:   customerclient.NewMockClient(ctrl),
		mockCart:       cartclient.NewMockClient(ctrl),
		mockPayment:    payment.NewMockClient(ctrl),
		mockWishlist:   wishlistclient.NewMockClient(ctrl),
		mockSalesforce: salesforceclient.NewMockClient(ctrl),
		mockAddress:    addressclient.NewMockClient(ctrl),
		mockProduct:    productclient.NewMockClient(ctrl),
		mockWebhook:    webhookclient.NewMockClient(ctrl),
	}
}

func newServiceUnderTest(tc *testController) *service {
	logger, _ := zap.NewDevelopment()
	return &service{
		repo:             tc.mockRepo,
		log:              logger.Sugar(),
		customerClient:   tc.mockCustomer,
		cartClient:       tc.mockCart,
		paymentClient:    tc.mockPayment,
		wishlistClient:   tc.mockWishlist,
		salesforceClient: tc.mockSalesforce,
		addressClient:    tc.mockAddress,
		productClient:    tc.mockProduct,
		webhookClient:    tc.mockWebhook,
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
			City:        stringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: stringPtr("1234567890"),
		}, nil)

	tc.mockCart.EXPECT().
		GetCart(gomock.Any()).
		Return(&cartEntities.Cart{
			Id:             cartID,
			TaxAmount:      decimal.NewFromFloat(10.0),
			TaxCurrency:    "USD",
			ShippingRateID: shippingRateID,
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
			StripeID: stringPtr(customerStripeID),
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

	req := &entities.CreateOrderRequest{
		Body: &entities.CreateOrderRequestBody{
			AddressID:             addressID,
			ShippingRateID:        shippingRateID,
			StripePaymentMethodID: paymentMethodID,
		},
	}

	resp, err := s.CreateOrder(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.OrderReference)
	// wait for async notify call to be done
	<-notifyCallDone
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

	ctx := sharedMeta.WithXCustomerID(context.Background(), customerID.String())

	tc.mockAddress.EXPECT().
		GetAddress(gomock.Any(), &addressEntities.GetAddressRequest{
			AddressID: addressID,
		}).
		Return(&addressEntities.Address{
			FullName:    "John Doe",
			Address:     expectedAddress,
			City:        stringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: stringPtr("1234567890"),
		}, nil)

	tc.mockCart.EXPECT().
		GetCart(gomock.Any()).
		Return(&cartEntities.Cart{
			Id:             cartID,
			TaxAmount:      decimal.NewFromInt(10),
			TaxCurrency:    "USD",
			ShippingRateID: shippingRateID,
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
			AuthorizeNetID: stringPtr(customerAuthorizeNetID),
		}, nil)

	tc.mockPayment.EXPECT().
		GetProvider().
		Return(providers.ProviderAuthorizeNet).Times(2)

	tc.mockPayment.EXPECT().
		CreatePayment(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, req authorizenetEntities.CreatePaymentTransactionRequest) {
			assert.Equal(t, expectedTotal, req.Amount)
			assert.Equal(t, customerAuthorizeNetID, req.ProfileID)
			assert.Equal(t, paymentNonce, req.PaymentNonce)
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

	req := &entities.CreateOrderRequest{
		Body: &entities.CreateOrderRequestBody{
			AddressID:      addressID,
			ShippingRateID: shippingRateID,
			PaymentNonce:   paymentNonce,
		},
	}

	resp, err := s.CreateOrder(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.OrderReference)
	// wait for async notify call to be done
	<-notifyCallDone
}

func stringPtr(s string) *string {
	return &s
}
