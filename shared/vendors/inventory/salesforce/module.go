package salesforce

import (
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	Config     inventory.Config
	HttpClient *http.Client
	Logger     *zap.SugaredLogger
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) (client.Client, error) {
	cache := cache.New()
	svc := service.New(p.Config, p.HttpClient, p.Logger, cache)

	client := client.NewClient(svc)

	return client, nil
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)
