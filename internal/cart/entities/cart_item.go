package entities

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"
	"time"

	"github.com/google/uuid"
)

// swagger:model CartItem
type CartItem struct {
	ID               uuid.UUID `json:"id" gorm:"column:id"`
	CartID           uuid.UUID `json:"-" gorm:"column:cart_id"`
	ProductVariantID uuid.UUID `json:"product_variant_id" gorm:"column:product_variant_id"`
	Quantity         int       `json:"quantity" gorm:"column:quantity"`
	CreatedAt        time.Time `json:"added_at" gorm:"column:created_at"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at"`
}

type CartItemDetail struct {
	ID               uuid.UUID        `json:"id" gorm:"column:id"`
	CartID           uuid.UUID        `json:"-" gorm:"column:cart_id"`
	SKU              string           `json:"sku" gorm:"column:sku"`
	Name             string           `json:"name" db:"name"`
	Description      *string          `json:"description" gorm:"column:description"`
	ImageURL         string           `json:"image_url" db:"image_url"`
	ProductID        uuid.UUID        `json:"product_id" gorm:"column:product_id"`
	ProductVariantID uuid.UUID        `json:"-" gorm:"column:product_variant_id"`
	Price            decimal.Decimal  `json:"price" gorm:"column:price"`
	Currency         string           `json:"currency" gorm:"column:currency"`
	Attributes       *json.JSON       `json:"attributes" db:"attributes"`
	Length           *decimal.Decimal `json:"-" gorm:"column:length"`
	Width            *decimal.Decimal `json:"-" gorm:"column:width"`
	Height           *decimal.Decimal `json:"-" gorm:"column:height"`
	Weight           *decimal.Decimal `json:"-" gorm:"column:weight"`
	StripeTaxCode    *string          `json:"-" db:"stripe_tax_code"`
	Quantity         int              `json:"quantity" gorm:"column:quantity"`
	CreatedAt        time.Time        `json:"added_at" gorm:"column:created_at"`
	UpdatedAt        time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

func (CartItem) TableName() string {
	return "cart_items"
}
