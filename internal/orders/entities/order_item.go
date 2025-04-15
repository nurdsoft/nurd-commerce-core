package entities

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID               uuid.UUID        `json:"id" gorm:"column:id;default:gen_random_uuid()"`
	OrderID          uuid.UUID        `json:"order_id" gorm:"column:order_id"`
	ProductVariantID uuid.UUID        `json:"product_variant_id" gorm:"column:product_variant_id"`
	SKU              string           `json:"sku" gorm:"column:sku"`
	Description      *string          `json:"description" gorm:"column:description"`
	ImageURL         string           `json:"image_url" gorm:"column:image_url"`
	Name             string           `json:"name" gorm:"column:name"`
	Length           *decimal.Decimal `json:"length" gorm:"column:length"`
	Width            *decimal.Decimal `json:"width" gorm:"column:width"`
	Height           *decimal.Decimal `json:"height" gorm:"column:height"`
	Weight           *decimal.Decimal `json:"weight" gorm:"column:weight"`
	Quantity         int              `json:"quantity" gorm:"column:quantity"`
	Price            decimal.Decimal  `json:"price" gorm:"column:price"`
	Attributes       *json.JSON       `json:"attributes" db:"attributes"`
	CreatedAt        time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time        `json:"updated_at" gorm:"column:updated_at"`
	SalesforceID     string           `json:"-" gorm:"column:salesforce_id"`
}

func (m *OrderItem) TableName() string {
	return "order_items"
}
