package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	dbErrors "github.com/nurdsoft/nurd-commerce-core/shared/db"
	sharedJSON "github.com/nurdsoft/nurd-commerce-core/shared/json"

	"github.com/google/uuid"
	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/cart/errors"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
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
	tx := r.gormDB.WithContext(ctx).Model(&entities.Order{}).Where("id = ?", orderID)

	// Handle fulfillment_metadata append using Postgres || operator
	if newMetaRaw, ok := details["fulfillment_metadata"]; ok {
		newMetaBytes, err := json.Marshal(newMetaRaw)
		if err != nil {
			return err
		}

		// Wrap using your custom JSON type
		mergedPatch := sharedJSON.JSON(newMetaBytes)

		// Normalize 'null'::jsonb to '{}' inline using CASE
		tx = tx.Update("fulfillment_metadata", gorm.Expr(`
			CASE
				WHEN fulfillment_metadata IS NULL OR fulfillment_metadata = 'null'::jsonb THEN '{}'::jsonb
				ELSE fulfillment_metadata
			END || ?
		`, mergedPatch))

		// Remove from generic update to avoid conflict
		delete(details, "fulfillment_metadata")
	}

	// Update other fields if any
	if len(details) > 0 {
		tx = tx.Updates(details)
	}

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return moduleErrors.NewAPIError("ORDER_NOT_FOUND")
	}

	// Fetch related order items
	var orderItems []*entities.OrderItem
	if err := r.gormDB.WithContext(ctx).
		Where("order_id = ?", orderID).
		Find(&orderItems).Error; err != nil {
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
