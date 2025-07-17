package inventory

import (
	"context"
	"errors"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/providers"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
)

type Client interface {
	CreateOrder(ctx context.Context, req entities.CreateInventoryOrderRequest) (any, error)
	UpdateOrderStatus(ctx context.Context, req entities.UpdateInventoryOrderStatusRequest) error
	GetProducts(ctx context.Context, req any) (any, error)
	GetProductByID(ctx context.Context, req any) (any, error)
	GetProvider() providers.ProviderType
}

type localClient struct {
	provider         providers.ProviderType
	salesforceClient salesforce.Client
}

func NewClient(provider providers.ProviderType, salesforceClient salesforce.Client) Client {
	return &localClient{provider: provider, salesforceClient: salesforceClient}
}

func (c *localClient) GetProvider() providers.ProviderType {
	return c.provider
}

func (c *localClient) CreateOrder(ctx context.Context, req entities.CreateInventoryOrderRequest) (any, error) {
	switch c.provider {
	case providers.ProviderNone:
		return nil, nil
	case providers.ProviderSalesforce:
		return c.salesforceClient.CreateOrder(ctx, req)
	case providers.ProviderPrintful:
		return nil, errors.New("not implemented")
	default:
		return nil, errors.New("provider not supported")
	}
}

func (c *localClient) UpdateOrderStatus(ctx context.Context, req entities.UpdateInventoryOrderStatusRequest) error {
	switch c.provider {
	case providers.ProviderNone:
		return nil
	case providers.ProviderSalesforce:
		return c.salesforceClient.UpdateOrderStatus(ctx, req)
	case providers.ProviderPrintful:
		return errors.New("not implemented")
	default:
		return errors.New("provider not supported")
	}
}

func (c *localClient) GetProducts(ctx context.Context, req any) (any, error) {
	switch c.provider {
	case providers.ProviderNone:
		return nil, nil
	case providers.ProviderSalesforce:
		return nil, errors.New("not implemented")
	case providers.ProviderPrintful:
		return nil, errors.New("not implemented")
	default:
		return nil, errors.New("provider not supported")
	}
}

func (c *localClient) GetProductByID(ctx context.Context, req any) (any, error) {
	switch c.provider {
	case providers.ProviderNone:
		return nil, nil
	case providers.ProviderSalesforce:
		return nil, errors.New("not implemented")
	case providers.ProviderPrintful:
		return nil, errors.New("not implemented")
	default:
		return nil, errors.New("provider not supported")
	}
}
