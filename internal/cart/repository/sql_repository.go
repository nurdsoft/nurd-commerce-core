package repository

import (
	"context"
	"time"

	dbErrors "github.com/nurdsoft/nurd-commerce-core/shared/db"

	"github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	errors "github.com/nurdsoft/nurd-commerce-core/internal/cart/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type sqlRepository struct {
	gormDB *gorm.DB
}

func (r *sqlRepository) BeginTransaction(ctx context.Context) (Transaction, error) {
	tx := r.gormDB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

func (r *sqlRepository) GetActiveCart(ctx context.Context, customerID string) (*entities.Cart, error) {
	cart := &entities.Cart{}
	err := r.gormDB.WithContext(ctx).Where("customer_id = ? AND status = ?", customerID, entities.Active).First(cart).Error
	if err != nil && dbErrors.IsNotFoundError(err) {
		return nil, nil
	}
	return cart, err
}

func (r *sqlRepository) CreateNewCart(ctx context.Context, tx Transaction, customerID string) (*entities.Cart, error) {
	newCart := entities.Cart{
		Id:         uuid.New(),
		CustomerID: uuid.MustParse(customerID),
		Status:     entities.Active,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := tx.WithContext(ctx).Create(&newCart).Error; err != nil {
		return nil, err
	}
	return &newCart, nil
}

func (r *sqlRepository) UpdateCartStatus(ctx context.Context, tx Transaction, cartID string, status string) error {
	dbCtx := r.gormDB.WithContext(ctx)
	if tx != nil {
		dbCtx = tx.WithContext(ctx)
	}
	return dbCtx.Model(&entities.Cart{}).
		Where("id = ?", cartID).
		Update("status", status).Error
}

func (r *sqlRepository) GetCartItem(ctx context.Context, cartID, productVariantID string) (*entities.CartItem, error) {
	var item entities.CartItem
	err := r.gormDB.WithContext(ctx).
		Where("cart_id = ? AND product_variant_id = ?", cartID, productVariantID).
		First(&item).Error
	if err != nil && dbErrors.IsNotFoundError(err) {
		return nil, nil
	}
	return &item, err
}

// AddCartItem adds a new item to the cart.
func (r *sqlRepository) AddCartItem(ctx context.Context, tx Transaction, cartId, productVariantID string, quantity int) (*entities.CartItem, error) {
	newItem := entities.CartItem{
		ID:               uuid.New(),
		CartID:           uuid.MustParse(cartId),
		ProductVariantID: uuid.MustParse(productVariantID),
		Quantity:         quantity,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := tx.WithContext(ctx).Create(&newItem).Error; err != nil {
		return nil, err
	}
	return &newItem, nil
}

func (r *sqlRepository) UpdateCartItem(ctx context.Context, tx Transaction, itemID string, quantity int) error {
	return tx.WithContext(ctx).
		Model(&entities.CartItem{}).
		Where("id = ?", itemID).
		Updates(map[string]interface{}{
			"quantity": quantity,
		}).Error
}

func (r *sqlRepository) GetCartItems(ctx context.Context, cartID string) ([]entities.CartItemDetail, error) {
	var items []entities.CartItemDetail
	err := r.gormDB.WithContext(ctx).
		Table("cart_items").
		Joins("JOIN product_variants ON cart_items.product_variant_id = product_variants.id").
		Where("cart_id = ?", cartID).
		Select("cart_items.id, cart_items.cart_id, product_variants.sku, product_variants.name, product_variants.product_id, cart_items.product_variant_id, " +
			" product_variants.price, product_variants.currency, product_variants.attributes, product_variants.length, product_variants.width, " +
			" product_variants.height, product_variants.weight, product_variants.stripe_tax_code, cart_items.quantity, product_variants.image_url, " +
			" product_variants.description, cart_items.created_at, cart_items.updated_at").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *sqlRepository) RemoveCartItem(ctx context.Context, cartID, itemID string) error {
	return r.gormDB.WithContext(ctx).
		Where("id = ? AND cart_id = ?", itemID, cartID).
		Delete(&entities.CartItem{}).Error
}

func (r *sqlRepository) CreateCartShippingRates(ctx context.Context, shippingRates []entities.CartShippingRate) error {
	return r.gormDB.WithContext(ctx).Create(&shippingRates).Error
}

func (r *sqlRepository) GetShippingRate(ctx context.Context, shippingRate uuid.UUID) (*entities.CartShippingRate, error) {
	var rate entities.CartShippingRate
	err := r.gormDB.WithContext(ctx).
		Where("id = ?", shippingRate).
		First(&rate).Error

	if err != nil {
		if dbErrors.IsNotFoundError(err) {
			return nil, errors.NewAPIError("CART_SHIPPING_RATE_NOT_FOUND")
		}
		return nil, err
	}

	return &rate, err
}

func (r *sqlRepository) UpdateCartShippingAndTaxRate(ctx context.Context, cartID string, shippingRateID uuid.UUID, taxAmount decimal.Decimal, taxCurrency string, taxBreakdown json.JSON) error {
	return r.gormDB.WithContext(ctx).
		Model(&entities.Cart{}).
		Where("id = ?", cartID).
		Updates(map[string]interface{}{
			"shipping_rate_id": shippingRateID,
			"tax_amount":       taxAmount,
			"tax_currency":     taxCurrency,
			"tax_breakdown":    taxBreakdown,
		}).Error
}

func (r *sqlRepository) GetCartByID(ctx context.Context, cartID uuid.UUID) (*entities.Cart, error) {
	var cart entities.Cart
	err := r.gormDB.WithContext(ctx).
		Where("id = ?", cartID).
		First(&cart).Error
	if err != nil && dbErrors.IsNotFoundError(err) {
		return nil, nil
	}
	return &cart, err
}
