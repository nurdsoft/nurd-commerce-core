package productclient

import (
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/repository"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// ModuleParams for productclient.
type ModuleParams struct {
	fx.In

	DB           *sql.DB
	GormDB       *gorm.DB
	HTTPServer   *httpTransport.Server
	APPTransport svcTransport.Client
	CommonConfig cfg.Config
	Logger       *zap.SugaredLogger
}

// NewModule
// nolint:gocritic
func NewClientModule(p ModuleParams) Client {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(repo, p.Logger, p.CommonConfig)

	client := NewClient(svc)

	return client
}

var (
	// ModuleClient for uber fx.
	ModuleClient = fx.Options(fx.Provide(NewClientModule))
)
