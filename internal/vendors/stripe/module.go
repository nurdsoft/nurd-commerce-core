package stripe

import (
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/client"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/config"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	Config config.Config
	Logger *zap.SugaredLogger
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) (client.Client, error) {
	svc, err := service.New(p.Config, p.Logger)
	if err != nil {
		return nil, err
	}

	client := client.NewClient(svc)

	return client, nil
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)
