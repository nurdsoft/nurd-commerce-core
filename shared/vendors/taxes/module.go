package taxes

import (
	fakeprovider "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/fakeprovider"
	stripe "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/client"
	stripeService "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/service"
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
	svc, err := stripeService.New(p.Config.Stripe, p.Logger)
	if err != nil {
		return nil, err
	}

	stripeClient := stripe.NewClient(svc)
	fakeClient := fakeprovider.NewClient()

	client := NewClient(p.Config.Provider, stripeClient, fakeClient)

	return client, nil
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)
