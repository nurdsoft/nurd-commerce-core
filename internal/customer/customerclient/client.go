package customerclient

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/service"
)

type Client interface {
	GetCustomer(ctx context.Context) (*entities.Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*entities.Customer, error)
	UpdateCustomerAuthorizeNetID(ctx context.Context, id string, externalID string) error
	UpdateCustomerStripeID(ctx context.Context, id string, externalID string) error
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) GetCustomer(ctx context.Context) (*entities.Customer, error) {
	return c.svc.GetCustomer(ctx)
}

func (c *localClient) GetCustomerByID(ctx context.Context, id string) (*entities.Customer, error) {
	return c.svc.GetCustomerByID(ctx, id)
}

func (c *localClient) UpdateCustomerAuthorizeNetID(ctx context.Context, id string, externalID string) error {
	return c.svc.UpdateCustomerAuthorizeNetID(ctx, id, externalID)
}

func (c *localClient) UpdateCustomerStripeID(ctx context.Context, id string, externalID string) error {
	return c.svc.UpdateCustomerStripeID(ctx, id, externalID)
}
