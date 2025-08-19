package client

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/entities"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestClient_CalculateTax(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.CalculateTaxRequest{
		ShippingAmount: decimal.NewFromInt(1000),
		FromAddress:    &entities.Address{City: "City", State: "State", PostalCode: "12345", Country: "US"},
		ToAddress:      entities.Address{City: "City", State: "State", PostalCode: "12345", Country: "US"},
		TaxItems:       []entities.TaxItem{{Price: decimal.NewFromInt(1000), Quantity: 1, TaxCode: "tax_code"}},
	}
	stripeReq := &stripeEntities.CalculateTaxRequest{
		ShippingAmount: req.ShippingAmount,
		FromAddress: &stripeEntities.Address{
			Line1:      req.FromAddress.Line1,
			City:       req.FromAddress.City,
			State:      req.FromAddress.State,
			PostalCode: req.FromAddress.PostalCode,
			Country:    req.FromAddress.Country,
		},
		ToAddress: stripeEntities.Address{
			Line1:      req.ToAddress.Line1,
			City:       req.ToAddress.City,
			State:      req.ToAddress.State,
			PostalCode: req.ToAddress.PostalCode,
			Country:    req.ToAddress.Country,
		},
		TaxItems: mapStripeTaxItems(req.TaxItems),
	}

	t.Run("Customer tax location invalid", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, stripeReq).Return(nil, &appErrors.APIError{Message: "Customer tax location is invalid"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Customer tax location is invalid", err.Error())
	})

	t.Run("Shipping address invalid", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, stripeReq).Return(nil, &appErrors.APIError{Message: "Shipping address is invalid"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Shipping address is invalid", err.Error())
	})

	t.Run("Invalid tax location", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, stripeReq).Return(nil, &appErrors.APIError{Message: "Invalid tax location"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Invalid tax location", err.Error())
	})

	t.Run("Tax ID invalid", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, stripeReq).Return(nil, &appErrors.APIError{Message: "Tax ID is invalid"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Tax ID is invalid", err.Error())
	})

	t.Run("Stripe tax inactive", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, stripeReq).Return(nil, &appErrors.APIError{Message: "Stripe Tax is inactive"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Stripe Tax is inactive", err.Error())
	})

	t.Run("Taxes calculation failed", func(t *testing.T) {
		mockService.EXPECT().CalculateTax(ctx, stripeReq).Return(nil, &appErrors.APIError{Message: "Unable to calculate taxes"})

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Unable to calculate taxes", err.Error())
	})

}
