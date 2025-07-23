package product

import (
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/service"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/endpoints"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/repository"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// ModuleParams for product.
type ModuleParams struct {
	fx.In

	DB              *sql.DB
	GormDB          *gorm.DB
	HTTPServer      *httpTransport.Server
	APPTransport    svcTransport.Client
	CommonConfig    cfg.Config
	Logger          *zap.SugaredLogger
	InventoryClient inventory.Client
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) error {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(repo, p.InventoryClient, p.Logger, p.CommonConfig)
	eps := endpoints.New(svc)

	http.RegisterTransport(p.HTTPServer, eps, p.APPTransport)

	return nil
}

var (
	// ModuleHttpAPI for uber fx.
	ModuleHttpAPI = fx.Options(fx.Invoke(NewModule))
)
