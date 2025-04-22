package entities

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"
	"time"

	"github.com/google/uuid"
)

// swagger:model GetProductVariantResponse
type ProductVariant struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	ProductID     uuid.UUID        `json:"product_id" db:"product_id"`
	SKU           string           `json:"sku" db:"sku"`
	Name          string           `json:"name" db:"name"`
	Description   *string          `json:"description" db:"description"`
	ImageURL      *string          `json:"image_url" db:"image_url"`
	Price         decimal.Decimal  `json:"price" gorm:"column:price"`
	Currency      string           `json:"currency" gorm:"column:currency"`
	Length        *decimal.Decimal `json:"length" gorm:"column:length"`
	Width         *decimal.Decimal `json:"width" gorm:"column:width"`
	Height        *decimal.Decimal `json:"height" gorm:"column:height"`
	Weight        *decimal.Decimal `json:"weight" gorm:"column:weight"`
	Attributes    *json.JSON       `json:"attributes" db:"attributes"`
	StripeTaxCode *string          `json:"stripe_tax_code" db:"stripe_tax_code"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time       `json:"updated_at" db:"updated_at"`
}

func (u *ProductVariant) TableName() string {
	return "product_variants"
}
