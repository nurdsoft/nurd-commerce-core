package repository

import (
	"context"
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, product *entities.Product) (*entities.Product, error)
	FindByID(ctx context.Context, id string) (*entities.Product, error)
	FindByIDs(ctx context.Context, ids []string) ([]entities.Product, error)
	Update(ctx context.Context, details map[string]interface{}, id string) error
	CreateVariant(ctx context.Context, variant *entities.ProductVariant) (*entities.ProductVariant, error)
	FindVariantBySKU(ctx context.Context, sku string) (*entities.ProductVariant, error)
	UpdateVariant(ctx context.Context, details map[string]interface{}, id string) error
	FindVariantByID(ctx context.Context, id string) (*entities.ProductVariant, error)
	ListVariants(ctx context.Context, req *entities.ListProductVariantsRequest) (*entities.ListProductVariantsResponse, error)
}

// New repository for product.
func New(db *sql.DB, gormDB *gorm.DB) Repository {
	repo := &sqlRepository{gormDB}
	return repo
}
