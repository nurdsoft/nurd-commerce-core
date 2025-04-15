package salesforce

import (
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/client"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/config"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	Config     config.Config
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
