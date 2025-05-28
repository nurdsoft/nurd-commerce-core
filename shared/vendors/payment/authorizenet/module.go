package authorizenet

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	Config payment.Config
	Logger *zap.SugaredLogger
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) (client.Client, error) {
	svc := service.New(p.Config.AuthorizeNet, p.Logger)
	authClient := client.NewClient(svc)

	return authClient, nil
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)
