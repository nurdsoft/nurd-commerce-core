package fakeprovider

import (
	"context"
	"time"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/entities"
	"github.com/shopspring/decimal"
)

type Client interface {
	ValidateAddress(ctx context.Context, address entities.Address) (*entities.Address, error)
	GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error)
}

func NewClient() Client {
	return &localClient{}
}

type localClient struct{}

func (c *localClient) ValidateAddress(ctx context.Context, address entities.Address) (*entities.Address, error) {
	return &address, nil
}

func (c *localClient) GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error) {
	return []entities.ShippingRate{
		{
			Amount:                decimal.NewFromInt(100),
			Currency:              "USD",
			CarrierName:           "Fake Carrier",
			CarrierCode:           "FAKE",
			ServiceType:           "Standard",
			ServiceCode:           "STANDARD",
			EstimatedDeliveryDate: time.Now().Add(time.Hour * 24 * 3),
			BusinessDaysInTransit: "3",
			CreatedAt:             time.Now(),
		},
	}, nil
}
