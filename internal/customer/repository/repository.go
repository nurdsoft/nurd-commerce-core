package repository

import (
	"context"
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, customer *entities.Customer) (*entities.Customer, error)
	FindByEmail(ctx context.Context, email string) (*entities.Customer, error)
	FindByID(ctx context.Context, id string) (*entities.Customer, error)
	Update(ctx context.Context, details map[string]interface{}, id string) error
}

// New repository for customer.
func New(db *sql.DB, gormDB *gorm.DB) Repository {
	repo := &sqlRepository{gormDB}
	return repo
}
