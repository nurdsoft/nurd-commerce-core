package payment

import (
	authorizenetClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/client"
	authorizenetService "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
	stripeClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/client"
	stripeService "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/service"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	Config Config
	Logger *zap.SugaredLogger
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) (Client, error) {
	switch p.Config.Provider {
	case providers.ProviderStripe, "":
		service, err := stripeService.New(p.Config.Stripe, p.Logger)
		if err != nil {
			return nil, err
		}

		client := stripeClient.NewClient(service)
		return client, nil
	case providers.ProviderAuthorizeNet:
		service := authorizenetService.New(p.Config.AuthorizeNet, p.Logger)
		client := authorizenetClient.NewClient(service)

		return client, nil
	default:
		return nil, errors.Errorf("unknown provider: %s", p.Config.Provider)
	}
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)
