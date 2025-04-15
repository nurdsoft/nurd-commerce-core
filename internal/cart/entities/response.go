package entities

import (
	"time"

	"github.com/shopspring/decimal"
)

// swagger:model GetCartItemsResponse
type GetCartItemsResponse struct {
	Items []CartItemDetail `json:"items"`
}

// swagger:model GetShippingRateResponse
type GetShippingRateResponse struct {
	Rates []CartShippingRate `json:"rates"`
}

type ShippingRate struct {
	Amount                float64   `json:"amount"`
	Currency              string    `json:"currency"`
	CarrierName           string    `json:"carrier"`
	CarrierCode           string    `json:"carrier_code"`
	EstimatedDeliveryDate time.Time `json:"estimated_delivery_date"`
	ServiceType           string    `json:"service_type"`
}

// swagger:model GetTaxRateResponse
type GetTaxRateResponse struct {
	// Tax rate
	Tax decimal.Decimal `json:"tax"`
	// Total amount. Total = Price + Tax + Shipping Rate
	Total decimal.Decimal `json:"total"`
	// Subtotal amount. Subtotal = Price
	Subtotal decimal.Decimal `json:"subtotal"`
	// Shipping Rate
	ShippingRate decimal.Decimal `json:"shipping_rate"`
	// Currency of the total amount
	Currency string `json:"currency"`
}
