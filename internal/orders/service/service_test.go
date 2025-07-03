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
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/nurdsoft/nurd-commerce-core/shared/nullable"
	salesforceclient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	salesforceEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
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
			City:        nullable.StringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: nullable.StringPtr("1234567890"),
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
			City:        nullable.StringPtr("New York"),
			StateCode:   "NY",
			CountryCode: "US",
			PostalCode:  "10001",
			PhoneNumber: nullable.StringPtr("1234567890"),
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
			AuthorizeNetID: nullable.StringPtr(customerAuthorizeNetID),
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
		tc.mockSalesforce.EXPECT().
			UpdateOrderStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *salesforceEntities.UpdateOrderRequest) {
				assert.Equal(t, salesforceID, req.AccountID)
				assert.Equal(t, salesforceID, req.OrderId)
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
		tc.mockSalesforce.EXPECT().
			UpdateOrderStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *salesforceEntities.UpdateOrderRequest) {
				assert.Equal(t, salesforceID, req.AccountID)
				assert.Equal(t, salesforceID, req.OrderId)
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
						ID:     itemID,
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

func TestRefundOrder(t *testing.T) {
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
				// Order should be marked as fully refunded since we're refunding all items
				assert.Equal(t, entities.Refunded, orderData["status"])
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
				assert.False(t, hasStatus, "Order status should not be changed for partial refund")

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

	t.Run("full refund changes order status", func(t *testing.T) {
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
		totalRefundAmount := decimal.NewFromInt(150) // 3 items * 50 each

		ctx := context.Background()

		existingOrder := &entities.Order{
			ID:                    orderID,
			CustomerID:            customerID,
			OrderReference:        orderRef,
			Status:                entities.PaymentSuccess,
			Total:                 decimal.NewFromInt(150),
			StripePaymentIntentID: &paymentIntentID,
		}

		// Total items: 3 (item1: 1, item2: 2), refunding all 3 items
		orderItems := []*entities.OrderItem{
			{
				ID:       orderItemID1,
				OrderID:  orderID,
				SKU:      itemSKU1,
				Price:    itemPrice,
				Quantity: 1,
				Status:   entities.ItemProcessing,
			},
			{
				ID:       orderItemID2,
				OrderID:  orderID,
				SKU:      itemSKU2,
				Price:    itemPrice,
				Quantity: 2,
				Status:   entities.ItemProcessing,
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
				assert.Equal(t, totalRefundAmount, req.Amount)
			}).
			Return(&providers.RefundResponse{
				ID:     refundID,
				Status: stripeEntities.StripeRefundSucceeded,
			}, nil)

		tc.mockRepo.EXPECT().
			UpdateOrderWithOrderItems(gomock.Any(), orderID, gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ uuid.UUID, orderData map[string]interface{}, itemsData map[string]interface{}) {
				// Order SHOULD be marked as fully refunded since we're refunding all 3 items
				status, hasStatus := orderData["status"]
				assert.True(t, hasStatus, "Order status should change for full refund")
				assert.Equal(t, entities.Refunded, status)

				// Check stripe refund total
				stripeRefundTotal, hasRefundTotal := orderData["stripe_refund_total"]
				assert.True(t, hasRefundTotal)
				assert.Equal(t, totalRefundAmount, stripeRefundTotal)

				// Check both items are marked for refund
				assert.Len(t, itemsData, 2)

				// Check item1 refund data
				itemData1, exists1 := itemsData[orderItemID1.String()]
				assert.True(t, exists1)
				itemMap1, ok1 := itemData1.(map[string]interface{})
				assert.True(t, ok1)
				assert.Equal(t, entities.ItemInitiatedRefund.String(), itemMap1["status"])
				assert.Equal(t, refundID, itemMap1["stripe_refund_id"])

				// Check item2 refund data
				itemData2, exists2 := itemsData[orderItemID2.String()]
				assert.True(t, exists2)
				itemMap2, ok2 := itemData2.(map[string]interface{})
				assert.True(t, ok2)
				assert.Equal(t, entities.ItemInitiatedRefund.String(), itemMap2["status"])
				assert.Equal(t, refundID, itemMap2["stripe_refund_id"])
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
						Quantity: 1,
					},
					{
						Sku:      itemSKU2,
						Quantity: 2,
					},
				},
			},
		}

		resp, err := s.RefundOrder(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, totalRefundAmount, resp.TotalRefundableAmount)
		assert.Len(t, resp.RefundableItems, 2)

		// Check first refundable item
		assert.Equal(t, orderItemID1.String(), resp.RefundableItems[0].ItemId)
		assert.Equal(t, itemSKU1, resp.RefundableItems[0].Sku)
		assert.Equal(t, 1, resp.RefundableItems[0].Quantity)
		assert.True(t, resp.RefundableItems[0].RefundInitiated)

		// Check second refundable item
		assert.Equal(t, orderItemID2.String(), resp.RefundableItems[1].ItemId)
		assert.Equal(t, itemSKU2, resp.RefundableItems[1].Sku)
		assert.Equal(t, 2, resp.RefundableItems[1].Quantity)
		assert.True(t, resp.RefundableItems[1].RefundInitiated)

		// wait for async notify call to be done
		<-notifyCallDone
	})
}
