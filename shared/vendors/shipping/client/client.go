package client

import (
	"context"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/entities"
)

type Client interface {
	ValidateAddress(ctx context.Context, address entities.Address) (*entities.Address, error)
	GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error)
}
