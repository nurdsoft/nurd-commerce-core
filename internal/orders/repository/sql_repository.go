package repository

import (
	"context"
	"encoding/base64"
	"time"

	dbErrors "github.com/nurdsoft/nurd-commerce-core/shared/db"

	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/cart/errors"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlRepository struct {
	gormDB *gorm.DB
}

func (r *sqlRepository) CreateOrder(ctx context.Context, cartID uuid.UUID, order *entities.Order, orderItems []*entities.OrderItem) error {
	// Start a transaction
	tx := r.gormDB.WithContext(ctx).Begin()

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return err
	}

	// mark cart items as purchased
	if err := tx.Model(&cartEntities.Cart{}).Where("id = ?", cartID).Update("status", cartEntities.Purchased).Error; err != nil {
		tx.Rollback()
		return err
	}

	// create order items in bulk
	if err := tx.Create(&orderItems).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *sqlRepository) ListOrders(ctx context.Context, customerID uuid.UUID, limit int, cursor string) ([]*entities.Order, string, error) {
	var orders []*entities.Order
	query := r.gormDB.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at DESC").Limit(limit + 1)

	if cursor != "" {
		decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return nil, "", err
		}
		query = query.Where("created_at < ?", string(decodedCursor))
	}

	if err := query.Find(&orders).Error; err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(orders) > limit {
		lastOrder := orders[limit-1]
		nextCursor = base64.StdEncoding.EncodeToString([]byte(lastOrder.CreatedAt.Format(time.RFC3339)))
		orders = orders[:limit] // Return only the number of records requested
	}

	return orders, nextCursor, nil
}

func (r *sqlRepository) GetOrderByStripePaymentIntentID(ctx context.Context, stripePaymentIntentID string) (*entities.Order, error) {
	order := &entities.Order{}
	if err := r.gormDB.WithContext(ctx).Where("stripe_payment_intent_id = ?", stripePaymentIntentID).First(order).Error; err != nil {
		return nil, err
	}

	return order, nil
}

func (r *sqlRepository) Update(ctx context.Context, details map[string]interface{}, orderID string, customerID string) error {
	result := r.gormDB.WithContext(ctx).Model(&entities.Order{}).Where("id = ?", orderID).Updates(details)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return moduleErrors.NewAPIError("ORDER_NOT_FOUND")
	}

	var orderItems []*entities.OrderItem
	if err := r.gormDB.WithContext(ctx).Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
		return moduleErrors.NewAPIError("ORDER_ERROR_GETTING_ITEMS")
	}
	return nil
}

func (r *sqlRepository) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*entities.Order, error) {
	order := &entities.Order{}
	if err := r.gormDB.WithContext(ctx).Where("id = ?", orderID).First(order).Error; err != nil {
		if dbErrors.IsNotFoundError(err) {
			return nil, moduleErrors.NewAPIError("ORDER_NOT_FOUND")
		}
		return nil, err
	}

	return order, nil
}

func (r *sqlRepository) GetOrderItemsByID(ctx context.Context, orderID uuid.UUID) ([]*entities.OrderItem, error) {
	var orderItems []*entities.OrderItem
	if err := r.gormDB.WithContext(ctx).Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
		return nil, err
	}

	return orderItems, nil
}

func (r *sqlRepository) AddSalesforceIDPerOrderItem(ctx context.Context, ids map[string]string) error {
	tx := r.gormDB.WithContext(ctx).Begin()
	for orderItemID, salesforceID := range ids {
		if err := tx.Model(&entities.OrderItem{}).Where("id = ?", orderItemID).Update("salesforce_id", salesforceID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *sqlRepository) OrderReferenceExists(ctx context.Context, orderReference string) (bool, error) {
	var count int64
	err := r.gormDB.WithContext(ctx).
		Model(&entities.Order{}).
		Where("order_reference = ?", orderReference).
		Count(&count).Error

	return count > 0, err
}

func (r *sqlRepository) GetOrderByReference(ctx context.Context, orderReference string) (*entities.Order, error) {
	order := &entities.Order{}
	if err := r.gormDB.WithContext(ctx).
		Where("order_reference = ?", orderReference).
		First(order).Error; err != nil {
		if dbErrors.IsNotFoundError(err) {
			return nil, moduleErrors.NewAPIError("ORDER_NOT_FOUND")
		}
		return nil, err
	}

	return order, nil
}
