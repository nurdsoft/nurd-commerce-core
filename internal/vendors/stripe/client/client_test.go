package client

import (
	"context"
	"errors"
	"testing"

	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/service"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v81"
)

func TestClient_CalculateTax(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.CalculateTaxRequest{
		ShippingAmount: decimal.NewFromInt(1000),
		FromAddress:    entities.Address{City: "City", State: "State", PostalCode: "12345", Country: "US"},
		ToAddress:      entities.Address{City: "City", State: "State", PostalCode: "12345", Country: "US"},
		TaxItems:       []entities.TaxItem{{Price: decimal.NewFromInt(1000), Quantity: 1, TaxCode: "tax_code"}},
	}

	t.Run("Customer tax location invalid", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, req).Return(nil, &appErrors.APIError{Message: "Customer tax location is invalid"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Customer tax location is invalid", err.Error())
	})

	t.Run("Shipping address invalid", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, req).Return(nil, &appErrors.APIError{Message: "Shipping address is invalid"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Shipping address is invalid", err.Error())
	})

	t.Run("Invalid tax location", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, req).Return(nil, &appErrors.APIError{Message: "Invalid tax location"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Invalid tax location", err.Error())
	})

	t.Run("Tax ID invalid", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, req).Return(nil, &appErrors.APIError{Message: "Tax ID is invalid"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Tax ID is invalid", err.Error())
	})

	t.Run("Stripe tax inactive", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, req).Return(nil, &appErrors.APIError{Message: "Stripe Tax is inactive"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Stripe Tax is inactive", err.Error())
	})

	t.Run("Taxes calculation failed", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, req).Return(nil, &appErrors.APIError{Message: "Unable to calculate taxes"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Unable to calculate taxes", err.Error())
	})

}

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
