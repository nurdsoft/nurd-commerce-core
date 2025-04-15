package client

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/shipengine/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/shipengine/service"
)

type Client interface {
	GetRatesEstimate(ctx context.Context, from, to entities.ShippingAddress, dimensions entities.Dimensions) ([]entities.EstimateRatesResponse, error)
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) GetRatesEstimate(ctx context.Context, from, to entities.ShippingAddress, dimensions entities.Dimensions) ([]entities.EstimateRatesResponse, error) {
	return c.svc.GetRatesEstimate(ctx, from, to, dimensions)
}
