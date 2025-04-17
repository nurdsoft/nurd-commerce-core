package addressclient

import (
	"database/sql"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/repository"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/service"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	shipengine "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ModuleParams for addressclient.
type ModuleParams struct {
	fx.In

	DB               *sql.DB
	GormDB           *gorm.DB
	HTTPServer       *httpTransport.Server
	APPTransport     svcTransport.Client
	CommonConfig     cfg.Config
	Logger           *zap.SugaredLogger
	ShipengineClient shipengine.Client
	SalesforceClient salesforce.Client
	CustomerClient   customerclient.Client
}

// NewClientModule
// nolint:gocritic
func NewClientModule(p ModuleParams) Client {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(repo, p.Logger, p.CommonConfig, p.ShipengineClient, p.SalesforceClient, p.CustomerClient)

	client := NewClient(svc)

	return client
}

var (
	// ModuleClient for uber fx.
	ModuleClient = fx.Options(fx.Provide(NewClientModule))
)
