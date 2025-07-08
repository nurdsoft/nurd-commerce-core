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
	ListOrders(ctx context.Context, customerID uuid.UUID, limit int, cursor string, includeItems bool) ([]*entities.Order, string, error)
	Update(ctx context.Context, details map[string]interface{}, orderID string, customerID string) error
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*entities.Order, error)
	GetOrderItemsByID(ctx context.Context, orderID uuid.UUID) ([]*entities.OrderItem, error)
	AddSalesforceIDPerOrderItem(ctx context.Context, ids map[string]string) error
	OrderReferenceExists(ctx context.Context, orderReference string) (bool, error)
	GetOrderByReference(ctx context.Context, orderReference string) (*entities.Order, error)
	GetOrderByStripePaymentIntentID(ctx context.Context, stripePaymentIntentID string) (*entities.Order, error)
	GetOrderByAuthorizeNetPaymentID(ctx context.Context, authorizeNetPaymentID string) (*entities.Order, error)
	UpdateOrderWithOrderItems(ctx context.Context, orderID uuid.UUID, orderData map[string]interface{}, orderItemsData map[string]interface{}) error
	GetOrderItemsByStripeRefundID(ctx context.Context, stripeRefundID string) ([]*entities.OrderItem, error)
}

func New(_ *sql.DB, gormDB *gorm.DB) Repository {
	repo := &sqlRepository{gormDB}
	return repo
}
