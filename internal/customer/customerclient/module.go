package customerclient

import (
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/repository"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/service"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	salesforce "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
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
	SalesforceClient salesforce.Client
}

// NewClientModule
// nolint:gocritic
func NewClientModule(p ModuleParams) Client {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(repo, p.Logger, p.CommonConfig, p.SalesforceClient)

	client := NewClient(svc)

	return client
}

var (
	// ModuleClient for uber fx.
	ModuleClient = fx.Options(fx.Provide(NewClientModule))
)
