package customerclient

import (
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/repository"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/service"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	salesforceclient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ModuleParams for customerclient.
type ModuleParams struct {
	fx.In

	DB               *sql.DB
	GormDB           *gorm.DB
	HTTPServer       *httpTransport.Server
	APPTransport     svcTransport.Client
	CommonConfig     cfg.Config
	Logger           *zap.SugaredLogger
	InventoryClient  inventory.Client
	SalesforceClient salesforceclient.Client
}

// NewClientModule
// nolint:gocritic
func NewClientModule(p ModuleParams) Client {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(repo, p.Logger, p.CommonConfig, p.SalesforceClient, p.InventoryClient)

	client := NewClient(svc)

	return client
}

var (
	// ModuleClient for uber fx.
	ModuleClient = fx.Options(fx.Provide(NewClientModule))
)
