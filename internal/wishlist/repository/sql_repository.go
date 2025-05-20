package repository

import (
	"context"
	"encoding/base64"
	"time"

	dbErrors "github.com/nurdsoft/nurd-commerce-core/shared/db"

	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/wishlist/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlRepository struct {
	gormDB *gorm.DB
}

func (r *sqlRepository) UpdateWishlist(ctx context.Context, customerID string, productIDs []uuid.UUID) error {
	var wishlists []entities.WishlistItem

	for _, productID := range productIDs {
		wishlists = append(wishlists, entities.WishlistItem{
			Id:         uuid.New(),
			CustomerID: uuid.MustParse(customerID),
			ProductID:  productID,
		})
	}

	result := r.gormDB.WithContext(ctx).Create(&wishlists)

	if result.Error != nil {
		// return success if product is already in wishlist
		if dbErrors.IsUniqueViolationError(result.Error) {
			return nil
		}
		return result.Error
	}

	return nil
}

func (r *sqlRepository) DeleteFromWishlist(ctx context.Context, customerID string, productID uuid.UUID) error {
	result := r.gormDB.WithContext(ctx).Where("customer_id = ? AND product_id = ?", customerID, productID).Delete(&entities.WishlistItem{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return moduleErrors.NewAPIError("WISHLIST_ITEM_NOT_FOUND")
	}

	return nil
}

func (r *sqlRepository) GetWishlist(ctx context.Context, customerID string, limit int, cursor string) ([]*entities.WishlistItem, string, error) {
	var wishlistItems []*entities.WishlistItem
	query := r.gormDB.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit + 1)
	}

	if cursor != "" {
		decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return nil, "", err
		}
		query = query.Where("created_at < ?", string(decodedCursor))
	}

	if err := query.Find(&wishlistItems).Error; err != nil {
		return nil, "", err
	}

	var nextCursor string
	if limit > 0 {
		if len(wishlistItems) > limit {
			lastItem := wishlistItems[limit-1]
			nextCursor = base64.StdEncoding.EncodeToString([]byte(lastItem.CreatedAt.Format(time.RFC3339)))
			wishlistItems = wishlistItems[:limit] // Return only the number of records requested
		}
	} else {
		nextCursor = ""
	}

	return wishlistItems, nextCursor, nil
}

func (r *sqlRepository) BulkRemoveFromWishlist(ctx context.Context, customerID uuid.UUID, productIDs []uuid.UUID) error {
	result := r.gormDB.WithContext(ctx).Where("customer_id = ? AND product_id IN ?", customerID, productIDs).Delete(&entities.WishlistItem{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *sqlRepository) GetWishlistProductTimestamps(customerID string, productIDs []uuid.UUID) (map[string]time.Time, error) {
	// map of product_id to the added_at time
	res := map[string]time.Time{}

	rows, err := r.gormDB.Table("wishlist_items").
		Select("product_id, created_at").
		Where("customer_id = ? AND product_id IN ?", customerID, productIDs).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var productID string
		var createdAt time.Time
		if err := rows.Scan(&productID, &createdAt); err != nil {
			return nil, err
		}
		res[productID] = createdAt
	}

	return res, nil
}
