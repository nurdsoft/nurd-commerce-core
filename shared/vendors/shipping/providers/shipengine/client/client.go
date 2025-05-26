package client

import (
	"context"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/shipengine/service"
)

type Client interface {
	ValidateAddress(ctx context.Context, address entities.Address) (*entities.Address, error)
	GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error)
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) ValidateAddress(ctx context.Context, address entities.Address) (*entities.Address, error) {
	return c.svc.ValidateAddress(ctx, address)
}

func (c *localClient) GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error) {
	return c.svc.GetShippingRates(ctx, shipment)
}
