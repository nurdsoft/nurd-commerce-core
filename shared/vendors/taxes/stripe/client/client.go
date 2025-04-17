package client

import (
	"context"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/service"
)

type Client interface {
	CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error)
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error) {
	return c.svc.CalculateTax(ctx, req)
}
