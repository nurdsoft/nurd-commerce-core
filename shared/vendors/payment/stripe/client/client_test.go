package client

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestClient_CreateCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.CreateCustomerRequest{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Phone: "1234567890",
	}

	t.Run("Success", func(t *testing.T) {
		expectedResp := &entities.CreateCustomerResponse{
			Id: "cus_123",
		}
		mockService.EXPECT().CreateCustomer(ctx, req).Return(expectedResp, nil)

		resp, err := client.CreateCustomer(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("Error", func(t *testing.T) {
		mockService.EXPECT().CreateCustomer(ctx, req).Return(nil, errors.New("service error"))

		resp, err := client.CreateCustomer(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "service error", err.Error())
	})
}

func TestClient_CreatePaymentIntent(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	customerId := "cus_123"
	// Create as a value, not a pointer
	req := entities.CreatePaymentIntentRequest{
		Amount:          decimal.NewFromInt(1000),
		Currency:        "usd",
		CustomerId:      &customerId,
		PaymentMethodId: "pm_123",
	}

	t.Run("Success", func(t *testing.T) {
		expectedResp := &entities.CreatePaymentIntentResponse{
			Id: "pi_123",
		}
		// When CreatePayment receives a value, it will pass a pointer to the service
		mockService.EXPECT().CreatePaymentIntent(ctx, gomock.Any()).Return(expectedResp, nil)

		resp, err := client.CreatePayment(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, "pi_123", resp.ID)
	})

	t.Run("Payment intent authentication failure", func(t *testing.T) {
		apiErr := &appErrors.APIError{Message: "Payment intent authentication failure"}
		mockService.EXPECT().CreatePaymentIntent(ctx, gomock.Any()).Return(nil, apiErr)

		resp, err := client.CreatePayment(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, apiErr, err)
		assert.Empty(t, resp.ID)
	})

	t.Run("Payment intent invalid parameter", func(t *testing.T) {
		apiErr := &appErrors.APIError{Message: "Payment intent invalid parameter"}
		mockService.EXPECT().CreatePaymentIntent(ctx, gomock.Any()).Return(nil, apiErr)

		resp, err := client.CreatePayment(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, apiErr, err)
		assert.Empty(t, resp.ID)
	})

	t.Run("Payment intent incompatible payment method", func(t *testing.T) {
		apiErr := &appErrors.APIError{Message: "Payment intent incompatible payment method"}
		mockService.EXPECT().CreatePaymentIntent(ctx, gomock.Any()).Return(nil, apiErr)

		resp, err := client.CreatePayment(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, apiErr, err)
		assert.Empty(t, resp.ID)
	})
}
