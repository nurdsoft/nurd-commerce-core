package wishlist

import (
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/service"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/endpoints"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/repository"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// ModuleParams for wishlist.
type ModuleParams struct {
	fx.In

	DB            *sql.DB
	GormDB        *gorm.DB
	HTTPServer    *httpTransport.Server
	APPTransport  svcTransport.Client
	CommonConfig  cfg.Config
	Logger        *zap.SugaredLogger
	ProductClient productclient.Client
	CartClient    cartclient.Client
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) error {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(repo, p.Logger, p.CommonConfig, p.ProductClient, p.CartClient)
	eps := endpoints.New(svc)

	http.RegisterTransport(p.HTTPServer, eps, p.APPTransport)

	return nil
}

var (
	// ModuleHttpAPI for uber fx.
	ModuleHttpAPI = fx.Options(fx.Invoke(NewModule))
)
