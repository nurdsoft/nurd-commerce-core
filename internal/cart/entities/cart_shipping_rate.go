package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CartShippingRate struct {
	Id                    uuid.UUID       `json:"id" gorm:"column:id"`
	CartID                uuid.UUID       `json:"-" gorm:"column:cart_id"`
	AddressID             uuid.UUID       `json:"-" gorm:"column:address_id"`
	Amount                decimal.Decimal `json:"amount" gorm:"column:amount"`
	Currency              string          `json:"currency" gorm:"column:currency"`
	CarrierName           string          `json:"carrier_name" gorm:"column:carrier_name"`
	CarrierCode           string          `json:"carrier_code" gorm:"column:carrier_code"`
	ServiceType           string          `json:"service_type" gorm:"column:service_type"`
	ServiceCode           string          `json:"service_code" gorm:"column:service_code"`
	EstimatedDeliveryDate time.Time       `json:"estimated_delivery_date" gorm:"column:estimated_delivery_date"`
	BusinessDaysInTransit string          `json:"business_days_in_transit" gorm:"column:business_days_in_transit"`
	CreatedAt             time.Time       `json:"-" gorm:"column:created_at"`
}

func (CartShippingRate) TableName() string {
	return "cart_shipping_rates"
}
