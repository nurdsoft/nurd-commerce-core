package repository

import (
	"context"
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	UpdateWishlist(ctx context.Context, customerID string, productIDs []uuid.UUID) error
	DeleteFromWishlist(ctx context.Context, customerID string, productID uuid.UUID) error
	GetWishlist(ctx context.Context, customerID string, limit int, cursor string) ([]*entities.WishlistItem, string, error)
	BulkRemoveFromWishlist(ctx context.Context, customerID uuid.UUID, productIDs []uuid.UUID) error
}

// New repository for wishlist.
func New(_ *sql.DB, gormDB *gorm.DB) Repository {
	repo := &sqlRepository{gormDB}
	return repo
}
