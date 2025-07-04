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

func (r *sqlRepository) ListOrders(ctx context.Context, customerID uuid.UUID, limit int, cursor string, includeItems bool) ([]*entities.Order, string, error) {
	// Base query for orders
	query := r.gormDB.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Limit(limit + 1)

	if cursor != "" {
		decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return nil, "", err
		}
		query = query.Where("created_at < ?", string(decodedCursor))
	}

	// Fetch orders
	var orders []*entities.Order
	if err := query.Find(&orders).Error; err != nil {
		return nil, "", err
	}

	// Handle pagination
	var nextCursor string
	if len(orders) > limit {
		lastOrder := orders[limit-1]
		nextCursor = base64.StdEncoding.EncodeToString([]byte(lastOrder.CreatedAt.Format(time.RFC3339)))
		orders = orders[:limit] // Return only the requested number of records
	}

	// If includeItems is true, fetch item summaries for all orders
	if includeItems && len(orders) > 0 {
		// Extract order IDs
		var orderIDs []uuid.UUID
		orderMap := make(map[uuid.UUID]*entities.Order)
		for _, order := range orders {
			orderIDs = append(orderIDs, order.ID)
			orderMap[order.ID] = order
			// Initialize the ItemSummary slice for each order
			order.ItemsSummary = make([]*entities.OrderItemSummary, 0)
		}

		// Query order items for all orders at once
		var orderItems []*entities.OrderItem
		err := r.gormDB.WithContext(ctx).
			Where("order_id IN ?", orderIDs).
			Find(&orderItems).Error

		if err != nil {
			return nil, "", err
		}

		// Convert order items to order item summaries and associate with orders
		for _, item := range orderItems {
			summary := &entities.OrderItemSummary{
				ID:               item.ID,
				ProductID:        item.ProductID,
				ProductVariantID: item.ProductVariantID,
				SKU:              item.SKU,
				ImageURL:         item.ImageURL,
				Name:             item.Name,
				Quantity:         item.Quantity,
				Price:            item.Price,
				Attributes:       item.Attributes,
				Status:           item.Status,
			}

			// Add summary to its parent order
			if order, exists := orderMap[item.OrderID]; exists {
				order.ItemsSummary = append(order.ItemsSummary, summary)
			}
		}
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

func (r *sqlRepository) GetOrderByAuthorizeNetPaymentID(ctx context.Context, authorizeNetPaymentID string) (*entities.Order, error) {
	order := &entities.Order{}
	if err := r.gormDB.WithContext(ctx).Where("authorizenet_payment_id = ?", authorizeNetPaymentID).First(order).Error; err != nil {
		return nil, err
	}

	return order, nil
}

func (r *sqlRepository) Update(ctx context.Context, details map[string]interface{}, orderID string, customerID string) error {
	tx := r.gormDB.Begin().WithContext(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update order items first if "items" key exists
	if items, ok := details["items"]; ok {
		itemsData, ok := items.([]map[string]interface{})
		if !ok {
			tx.Rollback()
			return moduleErrors.NewAPIError("INVALID_ITEMS_DATA_FORMAT")
		}

		// Update each order item individually
		for _, itemData := range itemsData {
			itemID, hasID := itemData["id"].(string)
			itemSKU, hasSKU := itemData["sku"].(string)

			// Validate that at least one identifier is provided
			if (!hasID || itemID == "") && (!hasSKU || itemSKU == "") {
				tx.Rollback()
				return moduleErrors.NewAPIError("INVALID_ITEM_IDENTIFIER")
			}

			// Remove ID and SKU from update data
			updateData := make(map[string]interface{})
			for k, v := range itemData {
				if k != "id" && k != "sku" {
					updateData[k] = v
				}
			}

			if len(updateData) > 0 {
				query := tx.Model(&entities.OrderItem{}).Where("order_id = ?", orderID)

				// Build WHERE clause based on available identifiers
				if hasID && itemID != "" && hasSKU && itemSKU != "" {
					// Both ID and SKU provided
					query = query.Where("(id = ? OR sku = ?)", itemID, itemSKU)
				} else if hasID && itemID != "" {
					// Only ID provided
					query = query.Where("id = ?", itemID)
				} else if hasSKU && itemSKU != "" {
					// Only SKU provided
					query = query.Where("sku = ?", itemSKU)
				}

				result := query.Updates(updateData)
				if result.Error != nil {
					tx.Rollback()
					return result.Error
				}
			}
		}

		// Remove from generic update to avoid conflict
		delete(details, "items")
	}

	// Handle fulfillment_metadata append using Postgres || operator
	if newMetaRaw, ok := details["fulfillment_metadata"]; ok {
		newMetaBytes, err := json.Marshal(newMetaRaw)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Wrap using your custom JSON type
		mergedPatch := sharedJSON.JSON(newMetaBytes)

		// Normalize 'null'::jsonb to '{}' inline using CASE
		result := tx.Model(&entities.Order{}).Where("id = ?", orderID).Update("fulfillment_metadata", gorm.Expr(`
			CASE
				WHEN fulfillment_metadata IS NULL OR fulfillment_metadata = 'null'::jsonb THEN '{}'::jsonb
				ELSE fulfillment_metadata
			END || ?
		`, mergedPatch))

		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		// Remove from generic update to avoid conflict
		delete(details, "fulfillment_metadata")
	}

	// Update other order fields if any
	if len(details) > 0 {
		result := tx.Model(&entities.Order{}).Where("id = ?", orderID).Updates(details)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			return moduleErrors.NewAPIError("ORDER_NOT_FOUND")
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
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

func (r *sqlRepository) UpdateOrderWithOrderItems(ctx context.Context, orderID uuid.UUID, orderData map[string]interface{}, orderItemsData map[string]interface{}) error {
	tx := r.gormDB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if orderData != nil {
		// Update order status
		if err := tx.Model(&entities.Order{}).Where("id = ?", orderID).Updates(orderData).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(orderItemsData) > 0 {
		// Update order items with refund data
		for itemID, data := range orderItemsData {
			if err := tx.Model(&entities.OrderItem{}).
				Where("id = ?", itemID).
				Where("order_id = ?", orderID).
				Updates(data).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

func (r *sqlRepository) GetOrderItemsByStripeRefundID(ctx context.Context, stripeRefundID string) ([]*entities.OrderItem, error) {
	var orderItems []*entities.OrderItem
	if err := r.gormDB.WithContext(ctx).
		Where("stripe_refund_id = ?", stripeRefundID).
		Find(&orderItems).Error; err != nil {
		return nil, err
	}

	return orderItems, nil
}
