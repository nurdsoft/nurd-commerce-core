package taxes

import (
	"context"
	"errors"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/entities"
	fakeprovider "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/fakeprovider"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/providers"
	stripe "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/client"
	taxjar "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/taxjar/client"
)

type Client interface {
	CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error)
	GetProvider() providers.ProviderType
}

type localClient struct {
	provider     providers.ProviderType
	stripeClient stripe.Client
	fakeClient   fakeprovider.Client
	taxjarClient taxjar.Client
}

func NewClient(provider providers.ProviderType, stripeClient stripe.Client, fakeClient fakeprovider.Client, taxjarClient taxjar.Client) Client {
	return &localClient{provider: provider, stripeClient: stripeClient, fakeClient: fakeClient, taxjarClient: taxjarClient}
}

func (c *localClient) CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error) {
	switch c.provider {
	case "", providers.ProviderNone:
		return c.fakeClient.CalculateTax(ctx, req)
	case providers.ProviderStripe:
		return c.stripeClient.CalculateTax(ctx, req)
	case providers.ProviderTaxJar:
		return c.taxjarClient.CalculateTax(ctx, req)
	default:
		return nil, errors.New("invalid tax provider")
	}
}

func (c *localClient) GetProvider() providers.ProviderType {
	return c.provider
}
