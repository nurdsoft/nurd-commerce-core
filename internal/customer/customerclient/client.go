package customerclient

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
)

type Client interface {
	GetCustomer(ctx context.Context) (*entities.Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*entities.Customer, error)
	UpdateCustomerExternalID(ctx context.Context, id string, externalID string, paymentProvider providers.ProviderType) error
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

func (c *localClient) UpdateCustomerExternalID(ctx context.Context, id string, externalID string, paymentProvider providers.ProviderType) error {
	return c.svc.UpdateCustomerExternalID(ctx, id, externalID, paymentProvider)
}
