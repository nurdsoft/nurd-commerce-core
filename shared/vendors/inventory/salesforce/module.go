package salesforce

import (
	"database/sql"
	"net/http"

	ordersrepo "github.com/nurdsoft/nurd-commerce-core/internal/orders/repository"
	productRepo "github.com/nurdsoft/nurd-commerce-core/internal/product/repository"
	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	Config     inventory.Config
	HttpClient *http.Client
	Logger     *zap.SugaredLogger
	DB         *sql.DB
	GormDB     *gorm.DB
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) (client.Client, error) {
	cache := cache.New()
	svc := service.New(p.Config.Salesforce, p.HttpClient, p.Logger, cache)

	ordersRepo := ordersrepo.New(p.DB, p.GormDB)
	productsRepo := productRepo.New(p.DB, p.GormDB)
	client := client.NewClient(svc, p.Config.Provider, productsRepo, ordersRepo)

	return client, nil
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)
