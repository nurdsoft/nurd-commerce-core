package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	customerEntities "github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/ordersclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/entities"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	stripeClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/client"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
)

func Test_service_GetPaymentMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (
		*service, context.Context,
		*customerclient.MockClient,
		*stripeClient.MockClient,
	) {
		mockCustomerClient := customerclient.NewMockClient(ctrl)
		mockStripeClient := stripeClient.NewMockClient(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			log:            zap.NewExample().Sugar(),
			stripeClient:   mockStripeClient,
			customerClient: mockCustomerClient,
		}
		return svc, ctx, mockCustomerClient, mockStripeClient
	}

	t.Run("Valid request with existing stripe customer", func(t *testing.T) {
		svc, ctx, mockCustomerClient, mockStripeClient := setup()
		stripeID := "cust_123"
		customer := &customerEntities.Customer{
			ID:       uuid.New(),
			StripeID: &stripeID,
		}

		mockCustomerClient.EXPECT().
			GetCustomerByID(ctx, meta.XCustomerID(ctx)).Return(customer, nil).Times(1)
		mockStripeClient.EXPECT().
			GetCustomerPaymentMethods(ctx, &stripeID).Return(&stripeEntities.GetCustomerPaymentMethodsResponse{
			PaymentMethods: []stripeEntities.PaymentMethod{
				{
					Id:    "pm_123",
					Brand: "visa",
					Last4: "4242",
				},
			},
		}, nil).Times(1)

		resp, err := svc.GetPaymentMethods(ctx)

		assert.NoError(t, err)
		assert.Len(t, resp.PaymentMethods, 1)
		assert.Equal(t, "pm_123", resp.PaymentMethods[0].Id)
	})

	t.Run("no customer ID", func(t *testing.T) {
		svc, _, _, _ := setup()
		ctx := meta.WithXCustomerID(context.Background(), "")

		_, err := svc.GetPaymentMethods(ctx)

		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_GetSetupIntent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (
		*service, context.Context,
		*customerclient.MockClient,
		*stripeClient.MockClient,
	) {
		mockCustomerClient := customerclient.NewMockClient(ctrl)
		mockStripeClient := stripeClient.NewMockClient(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			log:            zap.NewExample().Sugar(),
			stripeClient:   mockStripeClient,
			customerClient: mockCustomerClient,
		}
		return svc, ctx, mockCustomerClient, mockStripeClient
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockCustomerClient, mockStripeClient := setup()
		stripeID := "cust_123"
		customer := &customerEntities.Customer{
			ID:       uuid.New(),
			StripeID: &stripeID,
		}

		mockCustomerClient.EXPECT().
			GetCustomerByID(ctx, meta.XCustomerID(ctx)).Return(customer, nil).Times(1)
		mockStripeClient.EXPECT().
			GetSetupIntent(ctx, &stripeID).Return(&stripeEntities.GetSetupIntentResponse{
			SetupIntent:  "seti_123",
			EphemeralKey: "ephkey_123",
			CustomerId:   "cust_123",
		}, nil).Times(1)

		resp, err := svc.GetSetupIntent(ctx)

		assert.NoError(t, err)
		assert.Equal(t, "seti_123", resp.SetupIntent.SetupIntent)
	})

	t.Run("no customer ID", func(t *testing.T) {
		svc, _, _, _ := setup()
		ctx := meta.WithXCustomerID(context.Background(), "")

		_, err := svc.GetSetupIntent(ctx)

		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_GetPaymentMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (
		*service, context.Context,
		*customerclient.MockClient,
		*stripeClient.MockClient,
	) {
		mockCustomerClient := customerclient.NewMockClient(ctrl)
		mockStripeClient := stripeClient.NewMockClient(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			log:            zap.NewExample().Sugar(),
			stripeClient:   mockStripeClient,
			customerClient: mockCustomerClient,
		}
		return svc, ctx, mockCustomerClient, mockStripeClient
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockCustomerClient, mockStripeClient := setup()
		stripeID := "cust_123"
		paymentMethodID := "pm_123"
		customer := &customerEntities.Customer{
			ID:       uuid.New(),
			StripeID: &stripeID,
		}
		req := &entities.StripeGetPaymentMethodRequest{
			PaymentMethodId: paymentMethodID,
		}

		mockCustomerClient.EXPECT().
			GetCustomerByID(ctx, meta.XCustomerID(ctx)).Return(customer, nil).Times(1)
		mockStripeClient.EXPECT().
			GetCustomerPaymentMethodById(ctx, &stripeID, &paymentMethodID).Return(&stripeEntities.GetCustomerPaymentMethodResponse{
			PaymentMethod: stripeEntities.PaymentMethod{
				Id:    "pm_123",
				Brand: "visa",
				Last4: "4242",
			},
		}, nil).Times(1)

		resp, err := svc.GetPaymentMethod(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "pm_123", resp.PaymentMethod.Id)
	})

	t.Run("no customer ID", func(t *testing.T) {
		svc, _, _, _ := setup()
		ctx := meta.WithXCustomerID(context.Background(), "")
		req := &entities.StripeGetPaymentMethodRequest{
			PaymentMethodId: "pm_123",
		}

		_, err := svc.GetPaymentMethod(ctx, req)

		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("missing payment method ID", func(t *testing.T) {
		svc, ctx, mockCustomerClient, _ := setup()
		stripeID := "cust_123"
		customer := &customerEntities.Customer{
			ID:       uuid.New(),
			StripeID: &stripeID,
		}
		req := &entities.StripeGetPaymentMethodRequest{
			PaymentMethodId: "",
		}

		mockCustomerClient.EXPECT().
			GetCustomerByID(ctx, meta.XCustomerID(ctx)).Return(customer, nil).Times(1)

		_, err := svc.GetPaymentMethod(ctx, req)

		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_HandleStripeWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (
		*service, context.Context,
		*stripeClient.MockClient,
		*ordersclient.MockClient,
	) {
		mockStripeClient := stripeClient.NewMockClient(ctrl)
		mockOrdersClient := ordersclient.NewMockClient(ctrl)
		ctx := context.Background()
		svc := &service{
			log:          zap.NewExample().Sugar(),
			stripeClient: mockStripeClient,
			ordersClient: mockOrdersClient,
		}
		return svc, ctx, mockStripeClient, mockOrdersClient
	}

	t.Run("payment intent succeeded", func(t *testing.T) {
		svc, ctx, mockStripeClient, mockOrdersClient := setup()
		req := &entities.StripeWebhookRequest{
			Payload:   []byte("test_payload"),
			Signature: "test_signature",
		}

		mockStripeClient.EXPECT().
			GetWebhookEvent(ctx, gomock.Any()).Return(&stripeEntities.HandleWebhookEventResponse{
			Type:     "payment_intent.succeeded",
			ObjectId: "pi_123",
		}, nil).Times(1)
		mockOrdersClient.EXPECT().
			ProcessPaymentSucceeded(ctx, "pi_123").Return(nil).Times(1)

		err := svc.HandleStripeWebhook(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("payment intent failed", func(t *testing.T) {
		svc, ctx, mockStripeClient, mockOrdersClient := setup()
		req := &entities.StripeWebhookRequest{
			Payload:   []byte("test_payload"),
			Signature: "test_signature",
		}

		mockStripeClient.EXPECT().
			GetWebhookEvent(ctx, gomock.Any()).Return(&stripeEntities.HandleWebhookEventResponse{
			Type:     "payment_intent.payment_failed",
			ObjectId: "pi_123",
		}, nil).Times(1)
		mockOrdersClient.EXPECT().
			ProcessPaymentFailed(ctx, "pi_123").Return(nil).Times(1)

		err := svc.HandleStripeWebhook(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("unhandled event type", func(t *testing.T) {
		svc, ctx, mockStripeClient, _ := setup()
		req := &entities.StripeWebhookRequest{
			Payload:   []byte("test_payload"),
			Signature: "test_signature",
		}

		mockStripeClient.EXPECT().
			GetWebhookEvent(ctx, gomock.Any()).Return(&stripeEntities.HandleWebhookEventResponse{
			Type:     "customer.created",
			ObjectId: "cust_123",
		}, nil).Times(1)

		err := svc.HandleStripeWebhook(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("invalid signature", func(t *testing.T) {
		svc, ctx, mockStripeClient, _ := setup()
		req := &entities.StripeWebhookRequest{
			Payload:   []byte("test_payload"),
			Signature: "invalid_signature",
		}

		mockStripeClient.EXPECT().
			GetWebhookEvent(ctx, gomock.Any()).Return(nil, &appErrors.APIError{Message: "invalid signature"}).Times(1)

		err := svc.HandleStripeWebhook(ctx, req)

		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_getCustomerStripeID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (
		*service, context.Context,
		*customerclient.MockClient,
		*stripeClient.MockClient,
	) {
		mockCustomerClient := customerclient.NewMockClient(ctrl)
		mockStripeClient := stripeClient.NewMockClient(ctrl)
		ctx := context.Background()
		svc := &service{
			log:            zap.NewExample().Sugar(),
			stripeClient:   mockStripeClient,
			customerClient: mockCustomerClient,
		}
		return svc, ctx, mockCustomerClient, mockStripeClient
	}

	t.Run("customer with existing stripe ID", func(t *testing.T) {
		svc, ctx, mockCustomerClient, _ := setup()
		customerID := uuid.New().String()
		stripeID := "cust_123"
		customer := &customerEntities.Customer{
			ID:       uuid.New(),
			StripeID: &stripeID,
		}

		mockCustomerClient.EXPECT().
			GetCustomerByID(ctx, customerID).Return(customer, nil).Times(1)

		resultStripeID, created, err := svc.getCustomerStripeID(ctx, customerID)

		assert.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, "cust_123", *resultStripeID)
	})

	t.Run("customer without stripe ID creates new", func(t *testing.T) {
		svc, ctx, mockCustomerClient, mockStripeClient := setup()
		customerID := uuid.New().String()
		lastName := "Doe"
		customer := &customerEntities.Customer{
			ID:        uuid.New(),
			FirstName: "John",
			LastName:  &lastName,
			Email:     "john@example.com",
			StripeID:  nil,
		}

		mockCustomerClient.EXPECT().
			GetCustomerByID(ctx, customerID).Return(customer, nil).Times(1)
		mockStripeClient.EXPECT().
			CreateCustomer(ctx, gomock.Any()).Return(&stripeEntities.CreateCustomerResponse{
			Id: "cust_new123",
		}, nil).Times(1)
		mockCustomerClient.EXPECT().
			UpdateCustomerStripeID(ctx, customerID, "cust_new123").Return(nil).Times(1)

		resultStripeID, created, err := svc.getCustomerStripeID(ctx, customerID)

		assert.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, "cust_new123", *resultStripeID)
	})

	t.Run("customer not found", func(t *testing.T) {
		svc, ctx, mockCustomerClient, _ := setup()
		customerID := uuid.New().String()

		mockCustomerClient.EXPECT().
			GetCustomerByID(ctx, customerID).Return(nil, &appErrors.APIError{Message: "customer not found"}).Times(1)

		_, _, err := svc.getCustomerStripeID(ctx, customerID)

		assert.Error(t, err)
	})
}
