package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	"gorm.io/gorm"
)

type Repository interface {
	CreateOrder(ctx context.Context, cartID uuid.UUID, order *entities.Order, orderItems []*entities.OrderItem) error
	ListOrders(ctx context.Context, customerID uuid.UUID, limit int, cursor string) ([]*entities.Order, string, error)
	Update(ctx context.Context, details map[string]interface{}, orderID string, customerID string) error
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*entities.Order, error)
	GetOrderItemsByID(ctx context.Context, orderID uuid.UUID) ([]*entities.OrderItem, error)
	AddSalesforceIDPerOrderItem(ctx context.Context, ids map[string]string) error
	OrderReferenceExists(ctx context.Context, orderReference string) (bool, error)
	GetOrderByReference(ctx context.Context, orderReference string) (*entities.Order, error)
	GetOrderByExternalPaymentID(ctx context.Context, externalPaymentID string) (*entities.Order, error)
}

func New(_ *sql.DB, gormDB *gorm.DB) Repository {
	repo := &sqlRepository{gormDB}
	return repo
}
