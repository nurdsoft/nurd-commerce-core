package client

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/service"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v81"
)

func TestClient_CreateCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.CreatePaymentIntentRequest{
		Amount:          1000,
		Currency:        "usd",
		CustomerId:      stripe.String("cus_123"),
		PaymentMethodId: "pm_123",
	}

	t.Run("Payment intent authentication failure", func(t *testing.T) {
		mockService.EXPECT().CreatePaymentIntent(ctx, req).Return(nil, &appErrors.APIError{Message: "Payment intent authentication failure"})

		_, err := client.CreatePaymentIntent(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Payment intent authentication failure", err.Error())
	})

	t.Run("Payment intent invalid parameter", func(t *testing.T) {
		mockService.EXPECT().CreatePaymentIntent(ctx, req).Return(nil, &appErrors.APIError{Message: "Payment intent invalid parameter"})

		_, err := client.CreatePaymentIntent(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Payment intent invalid parameter", err.Error())
	})

	t.Run("Payment intent incompatible payment method", func(t *testing.T) {
		mockService.EXPECT().CreatePaymentIntent(ctx, req).Return(nil, &appErrors.APIError{Message: "Payment intent incompatible payment method"})

		_, err := client.CreatePaymentIntent(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Payment intent incompatible payment method", err.Error())
	})

}
