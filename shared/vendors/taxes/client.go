package taxes

import (
	"context"
	"errors"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/entities"
	fakeprovider "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/fakeprovider"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/providers"
	stripe "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/client"
)

type Client interface {
	CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error)
}

type localClient struct {
	provider     providers.ProviderType
	stripeClient stripe.Client
	fakeClient   fakeprovider.Client
}

func NewClient(provider providers.ProviderType, stripeClient stripe.Client, fakeClient fakeprovider.Client) Client {
	return &localClient{provider: provider, stripeClient: stripeClient, fakeClient: fakeClient}
}

func (c *localClient) CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error) {
	switch c.provider {
	case "":
		return c.fakeClient.CalculateTax(ctx, req)
	case providers.ProviderStripe:
		return c.stripeClient.CalculateTax(ctx, req)
	default:
		return nil, errors.New("invalid tax provider")
	}
}
