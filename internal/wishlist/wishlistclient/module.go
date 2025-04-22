package wishlistclient

import (
	"database/sql"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/repository"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// ModuleParams for wishlistclient.
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

// NewClientModule
// nolint:gocritic
func NewClientModule(p ModuleParams) Client {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(repo, p.Logger, p.CommonConfig, p.ProductClient, p.CartClient)

	client := NewClient(svc)

	return client
}

var (
	// ModuleClient for uber fx.
	ModuleClient = fx.Options(fx.Provide(NewClientModule))
)
