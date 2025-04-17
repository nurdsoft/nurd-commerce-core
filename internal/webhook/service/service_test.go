package service

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	orderEntities "github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	orders "github.com/nurdsoft/nurd-commerce-core/internal/orders/ordersclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/client"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/entities"
	webhookEntities "github.com/nurdsoft/nurd-commerce-core/internal/webhook/entities"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
)

func Test_service_HandleStripeWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, *client.MockClient, *orders.MockClient) {
		mockStripeClient := client.NewMockClient(ctrl)
		mockOrdersClient := orders.NewMockClient(ctrl)
		svc := &service{
			log:          zap.NewExample().Sugar(),
			stripeClient: mockStripeClient,
			ordersClient: mockOrdersClient,
		}
		return svc, mockStripeClient, mockOrdersClient
	}

	t.Run("Webhook signature verification fails", func(t *testing.T) {
		svc, mockStripeClient, _ := setup()
		ctx := context.Background()
		req := &webhookEntities.StripeWebhookRequest{
			Payload:   []byte("invalid_payload"),
			Signature: "invalid_signature",
		}

		mockStripeClient.EXPECT().GetWebhookEvent(ctx, gomock.Any()).Return(nil, &appErrors.APIError{Message: "Invalid signature"})

		err := svc.HandleStripeWebhook(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Contains(t, err.Error(), "Webhook signature verification failed")
	})

	t.Run("Unhandled event type", func(t *testing.T) {
		svc, mockStripeClient, _ := setup()
		ctx := context.Background()
		req := &webhookEntities.StripeWebhookRequest{
			Payload:   []byte("valid_payload"),
			Signature: "valid_signature",
		}

		event := &stripeEntities.HandleWebhookEventResponse{
			Type:     "unknown_event",
			ObjectId: "unknown_id",
		}

		mockStripeClient.EXPECT().GetWebhookEvent(ctx, gomock.Any()).Return(event, nil)

		err := svc.HandleStripeWebhook(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("payment_intent.succeeded - success", func(t *testing.T) {
		svc, mockStripeClient, mockOrdersClient := setup()
		ctx := context.Background()
		req := &webhookEntities.StripeWebhookRequest{
			Payload:   []byte("valid_payload"),
			Signature: "valid_signature",
		}

		event := &stripeEntities.HandleWebhookEventResponse{
			Type:     "payment_intent.succeeded",
			ObjectId: "pi_123",
		}

		mockStripeClient.EXPECT().GetWebhookEvent(ctx, gomock.Any()).Return(event, nil)
		mockOrdersClient.EXPECT().ProcessPaymentSucceeded(ctx, "pi_123").Return(nil)

		err := svc.HandleStripeWebhook(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("payment_intent.succeeded - failure in processing", func(t *testing.T) {
		svc, mockStripeClient, mockOrdersClient := setup()
		ctx := context.Background()
		req := &webhookEntities.StripeWebhookRequest{
			Payload:   []byte("valid_payload"),
			Signature: "valid_signature",
		}

		event := &stripeEntities.HandleWebhookEventResponse{
			Type:     "payment_intent.succeeded",
			ObjectId: "pi_123",
		}

		mockStripeClient.EXPECT().GetWebhookEvent(ctx, gomock.Any()).Return(event, nil)
		mockOrdersClient.EXPECT().ProcessPaymentSucceeded(ctx, "pi_123").Return(&appErrors.APIError{Message: "Order processing failed"})

		err := svc.HandleStripeWebhook(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("payment_intent.payment_failed - success", func(t *testing.T) {
		svc, mockStripeClient, mockOrdersClient := setup()
		ctx := context.Background()
		req := &webhookEntities.StripeWebhookRequest{
			Payload:   []byte("valid_payload"),
			Signature: "valid_signature",
		}

		event := &stripeEntities.HandleWebhookEventResponse{
			Type:     "payment_intent.payment_failed",
			ObjectId: "pi_123",
		}

		mockStripeClient.EXPECT().GetWebhookEvent(ctx, gomock.Any()).Return(event, nil)
		mockOrdersClient.EXPECT().ProcessPaymentFailed(ctx, "pi_123").Return(nil)

		err := svc.HandleStripeWebhook(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("payment_intent.payment_failed - failure in processing", func(t *testing.T) {
		svc, mockStripeClient, mockOrdersClient := setup()
		ctx := context.Background()
		req := &webhookEntities.StripeWebhookRequest{
			Payload:   []byte("valid_payload"),
			Signature: "valid_signature",
		}

		event := &stripeEntities.HandleWebhookEventResponse{
			Type:     "payment_intent.payment_failed",
			ObjectId: "pi_123",
		}

		mockStripeClient.EXPECT().GetWebhookEvent(ctx, gomock.Any()).Return(event, nil)
		mockOrdersClient.EXPECT().ProcessPaymentFailed(ctx, "pi_123").Return(&appErrors.APIError{Message: "Order processing failed"})

		err := svc.HandleStripeWebhook(ctx, req)
		assert.NoError(t, err)
	})
}

func Test_service_HandleIngramOrderUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, *orders.MockClient) {
		mockOrdersClient := orders.NewMockClient(ctrl)
		svc := &service{
			log:          zap.NewExample().Sugar(),
			ordersClient: mockOrdersClient,
		}
		return svc, mockOrdersClient
	}

	t.Run("success", func(t *testing.T) {
		svc, mockOrdersClient := setup()
		ctx := context.Background()
		req := &webhookEntities.IngramOrderUpdateRequest{
			Authorization: "",
			Payload: &webhookEntities.IngramOrderUpdateRequestBody{
				ClientOrderID:        gofakeit.Word(),
				OrderStatusCode:      "00",
				OrderSubtotal:        0,
				OrderDiscountAmount:  0,
				SalesTax:             0,
				ShippingHandling:     0,
				OrderTotal:           0,
				FreightCharge:        0,
				TotalItemDetailCount: 0,
				ShipmentDate:         "20250501",
				ConsumerPONumber:     "",
				Items:                []webhookEntities.Item{},
			},
		}

		ti, _ := time.Parse("20060102", req.Payload.ShipmentDate)

		expectedFullfillment := &orderEntities.IngramOrderFullfillment{
			ShipmentDate:  ti,
			FreightCharge: decimal.NewFromInt(int64(req.Payload.FreightCharge)),
			OrderTotal:    decimal.NewFromInt(int64(req.Payload.OrderTotal)),
		}

		mockOrdersClient.EXPECT().ProcessOrderShipped(ctx, req.Payload.ClientOrderID, expectedFullfillment).Return(nil)

		err := svc.HandleIngramOrderUpdate(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("failure in processing", func(t *testing.T) {
		svc, mockOrdersClient := setup()
		ctx := context.Background()
		orderRef := gofakeit.Word()
		req := &webhookEntities.IngramOrderUpdateRequest{
			Authorization: "",
			Payload: &webhookEntities.IngramOrderUpdateRequestBody{
				ClientOrderID:        orderRef,
				OrderStatusCode:      "00",
				OrderSubtotal:        0,
				OrderDiscountAmount:  0,
				SalesTax:             0,
				ShippingHandling:     0,
				OrderTotal:           0,
				FreightCharge:        0,
				TotalItemDetailCount: 0,
				ShipmentDate:         "20250501",
				ConsumerPONumber:     "",
				Items:                []webhookEntities.Item{},
			},
		}

		ti, _ := time.Parse("20060102", req.Payload.ShipmentDate)

		expectedFullfillment := &orderEntities.IngramOrderFullfillment{
			ShipmentDate:  ti,
			FreightCharge: decimal.NewFromInt(int64(req.Payload.FreightCharge)),
			OrderTotal:    decimal.NewFromInt(int64(req.Payload.OrderTotal)),
		}

		mockOrdersClient.EXPECT().ProcessOrderShipped(ctx, orderRef, expectedFullfillment).Return(&appErrors.APIError{Message: "Order processing failed"})

		err := svc.HandleIngramOrderUpdate(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_HandleIngramInvoiceUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, *orders.MockClient) {
		mockOrdersClient := orders.NewMockClient(ctrl)
		svc := &service{
			log:          zap.NewExample().Sugar(),
			ordersClient: mockOrdersClient,
		}
		return svc, mockOrdersClient
	}

	t.Run("success", func(t *testing.T) {
		svc, mockOrdersClient := setup()
		ctx := context.Background()
		orderRef := gofakeit.Word()
		req := &webhookEntities.IngramOrderInvoiceRequest{
			Authorization: "",
			Payload: []webhookEntities.IngramOrderInvoiceRequestBody{
				{
					ClientOrderID:      orderRef,
					InvoiceNumber:      "",
					WarehouseSAN:       "",
					InvoiceDate:        "",
					TotalNetPrice:      0,
					TotalShipping:      0,
					TotalHandling:      0,
					TotalGiftWrap:      0,
					TotalInvoice:       0,
					TotalInvoiceWeight: 0,
					BillLadingNumber:   "",
					AmountDue:          0,
					Items:              []webhookEntities.InvoiceItem{},
				},
				{
					ClientOrderID:      orderRef,
					InvoiceNumber:      "",
					WarehouseSAN:       "",
					InvoiceDate:        "",
					TotalNetPrice:      0,
					TotalShipping:      0,
					TotalHandling:      0,
					TotalGiftWrap:      0,
					TotalInvoice:       0,
					TotalInvoiceWeight: 0,
					BillLadingNumber:   "",
					AmountDue:          0,
					Items:              []webhookEntities.InvoiceItem{},
				},
			},
		}

		mockOrdersClient.EXPECT().ProcessOrderInvoice(ctx, orderRef, gomock.Any()).Return(nil)

		err := svc.HandleIngramInvoiceUpdate(ctx, req)
		assert.NoError(t, err)
	})
}
