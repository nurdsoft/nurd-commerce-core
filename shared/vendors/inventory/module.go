package inventory

import (
	salesforceclient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	Config           Config
	Logger           *zap.SugaredLogger
	SalesforceClient salesforceclient.Client
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) (Client, error) {
	client := NewClient(p.Config.Provider, p.SalesforceClient)

	return client, nil
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)
