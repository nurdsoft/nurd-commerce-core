package inventory

import (
	"context"
	"errors"

	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/entities"
	printful "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/providers"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
)

type Client interface {
	CreateOrder(ctx context.Context, req entities.CreateInventoryOrderRequest) (any, error)
	UpdateOrderStatus(ctx context.Context, req entities.UpdateInventoryOrderStatusRequest) error
	GetProducts(ctx context.Context, req entities.ListProductsRequest) (entities.ListProductsResponse, error)
	GetProductByID(ctx context.Context, id string) (*entities.Product, error)
	GetProvider() providers.ProviderType
}

type localClient struct {
	provider         providers.ProviderType
	salesforceClient salesforce.Client
	printfulClient   printful.Client
}

func NewClient(provider providers.ProviderType, salesforceClient salesforce.Client, printfulClient printful.Client) Client {
	return &localClient{provider: provider, salesforceClient: salesforceClient, printfulClient: printfulClient}
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

func (c *localClient) GetProducts(ctx context.Context, req entities.ListProductsRequest) (entities.ListProductsResponse, error) {
	switch c.provider {
	case providers.ProviderNone:
		return entities.ListProductsResponse{}, nil
	case providers.ProviderSalesforce:
		return entities.ListProductsResponse{}, appErrors.NewAPIError("PROVIDER_NOT_IMPLEMENTED", "GetProducts is not implemented for Salesforce")
	case providers.ProviderPrintful:
		return c.printfulClient.GetSyncProducts(ctx, req)
	default:
		return entities.ListProductsResponse{}, errors.New("provider not supported")
	}
}

func (c *localClient) GetProductByID(ctx context.Context, id string) (*entities.Product, error) {
	switch c.provider {
	case providers.ProviderNone:
		return nil, nil
	case providers.ProviderSalesforce:
		return c.salesforceClient.GetProductByID(ctx, id)
	case providers.ProviderPrintful:
		return c.printfulClient.GetSyncProduct(ctx, id)
	default:
		return nil, errors.New("provider not supported")
	}
}
