package client

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/shipengine/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetRatesEstimate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	shipment := entities.Shipment{
		Origin:      entities.Address{CountryCode: "US", PostalCode: "37086", City: "La Vergne", StateCode: "TN"},
		Destination: entities.Address{CountryCode: "US", PostalCode: "12345", City: "City", StateCode: "State"},
		Dimensions:  entities.Dimensions{Length: decimal.NewFromInt(10), Width: decimal.NewFromInt(10), Height: decimal.NewFromInt(10), Weight: decimal.NewFromInt(10)},
	}

	t.Run("Invalid delivery address postal code", func(t *testing.T) {
		mockService.EXPECT().GetShippingRates(ctx, shipment).Return(nil, &appErrors.APIError{Message: "Invalid delivery address postal code"})

		_, err := client.GetShippingRates(ctx, shipment)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Invalid delivery address postal code", err.Error())
	})

	t.Run("Invalid origin address postal code", func(t *testing.T) {
		mockService.EXPECT().GetShippingRates(ctx, shipment).Return(nil, &appErrors.APIError{Message: "Invalid origin address postal code"})

		_, err := client.GetShippingRates(ctx, shipment)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Invalid origin address postal code", err.Error())
	})

	t.Run("Missing carrier while getting shipping estimates", func(t *testing.T) {
		mockService.EXPECT().GetShippingRates(ctx, shipment).Return(nil, &appErrors.APIError{Message: "Missing carrier while getting shipping estimates"})

		_, err := client.GetShippingRates(ctx, shipment)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Missing carrier while getting shipping estimates", err.Error())
	})

	t.Run("Package weight exceeds carrier limit", func(t *testing.T) {
		mockService.EXPECT().GetShippingRates(ctx, shipment).Return(nil, &appErrors.APIError{Message: "Package weight exceeds carrier limit"})

		_, err := client.GetShippingRates(ctx, shipment)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Package weight exceeds carrier limit", err.Error())
	})

	t.Run("Carrier error", func(t *testing.T) {
		mockService.EXPECT().GetShippingRates(ctx, shipment).Return(nil, &appErrors.APIError{Message: "Carrier error"})

		_, err := client.GetShippingRates(ctx, shipment)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Carrier error", err.Error())
	})
}
