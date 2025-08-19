package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"
)

type CartStatus string

const (
	Active    CartStatus = "active"
	Purchased CartStatus = "purchased"
	Cleared   CartStatus = "cleared"
)

type Cart struct {
	Id           uuid.UUID       `json:"id" gorm:"column:id"`
	CustomerID   uuid.UUID       `json:"customer_id" gorm:"column:customer_id"`
	Status       CartStatus      `db:"cart_status"`
	TaxAmount    decimal.Decimal `json:"tax_amount" gorm:"column:tax_amount"`
	TaxCurrency  string          `json:"tax_currency" gorm:"column:tax_currency"`
	TaxBreakdown json.JSON       `json:"tax_breakdown" gorm:"column:tax_breakdown"`
	CreatedAt    time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (Cart) TableName() string {
	return "carts"
}
