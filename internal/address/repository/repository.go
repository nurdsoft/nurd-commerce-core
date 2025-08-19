package repository

import (
	"context"
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	"gorm.io/gorm"
)

type Repository interface {
	CreateAddress(ctx context.Context, address *entities.Address) (*entities.Address, error)
	GetAddresses(ctx context.Context, customerID string) ([]entities.Address, error)
	GetAddress(ctx context.Context, customerID, addressID string) (*entities.Address, error)
	UpdateAddress(ctx context.Context, address *entities.Address) (*entities.Address, error)
	DeleteAddress(ctx context.Context, customerID, addressID string) error
	UpdateAddressField(ctx context.Context, customerID, addressID string, details map[string]interface{}) error
	GetDefaultAddress(ctx context.Context, customerID string) (*entities.Address, error)
}

// New repository for customer.
func New(db *sql.DB, gormDB *gorm.DB) Repository {
	repo := &sqlRepository{gormDB}
	return repo
}
