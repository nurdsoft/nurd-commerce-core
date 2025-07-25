package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"

	"gorm.io/gorm"
)

type Transaction interface {
	Commit() *gorm.DB
	Rollback() *gorm.DB
	WithContext(ctx context.Context) *gorm.DB
}

type Repository interface {
	BeginTransaction(ctx context.Context) (Transaction, error)
	GetActiveCart(ctx context.Context, customerID string) (*entities.Cart, error)
	CreateNewCart(ctx context.Context, tx Transaction, customerID string) (*entities.Cart, error)
	UpdateCartStatus(ctx context.Context, tx Transaction, cartID string, status string) error
	GetCartItem(ctx context.Context, cartID, productVariantID string) (*entities.CartItem, error)
	AddCartItem(ctx context.Context, tx Transaction, cartId, productVariantID string, quantity int) (*entities.CartItem, error)
	UpdateCartItem(ctx context.Context, tx Transaction, itemID string, quantity int) error
	GetCartItems(ctx context.Context, cartID string) ([]entities.CartItemDetail, error)
	RemoveCartItem(ctx context.Context, cartID, itemID string) error
	CreateCartShippingRates(ctx context.Context, shippingRate []entities.CartShippingRate) error
	GetShippingRate(ctx context.Context, shippingRateID uuid.UUID) (*entities.CartShippingRate, error)
	UpdateCartShippingAndTaxRate(ctx context.Context, cartID string, shippingRateId *uuid.UUID, taxAmount decimal.Decimal, taxCurrency string, taxBreakdown json.JSON) error
	GetCartByID(ctx context.Context, cartID uuid.UUID) (*entities.Cart, error)
}

func New(_ *sql.DB, gormDB *gorm.DB) Repository {
	repo := &sqlRepository{gormDB}
	return repo
}
