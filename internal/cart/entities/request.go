package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"
)

// swagger:parameters carts UpdateCartItemRequest
type UpdateCartItemRequest struct {
	// Item to be added to cart
	//
	// required: true
	// in:body
	Item *UpdateCartItemRequestBody
}

type UpdateCartItemRequestBody struct {
	ProductID   uuid.UUID           `json:"product_id"`
	Quantity    int                 `json:"quantity"`
	SKU         string              `json:"sku"`
	ProductData *ProductVariantData `json:"data"`
}

type ProductVariantData struct {
	Name          string           `json:"name"`
	Description   *string          `json:"description"`
	ImageURL      *string          `json:"image_url"`
	Price         decimal.Decimal  `json:"price"`
	Currency      string           `json:"currency" `
	Length        *decimal.Decimal `json:"length"`
	Width         *decimal.Decimal `json:"width"`
	Height        *decimal.Decimal `json:"height"`
	Weight        *decimal.Decimal `json:"weight"`
	Attributes    *json.JSON       `json:"attributes"`
	StripeTaxCode *string          `json:"stripe_tax_code"`
}

// swagger:parameters cart GetShippingRateRequest
type GetShippingRateRequest struct {
	// Request body
	//
	// required: true
	// in:body
	Body *GetShippingRateRequestBody
}

type GetShippingRateRequestBody struct {
	// Shipping Address UUID
	//
	// required: true
	// in:body
	AddressID uuid.UUID `json:"address_id"`
	// Warehouse address
	//
	// required: true
	// in:body
	// example:
	WarehouseAddress WarehouseAddress `json:"warehouse_address"`
	// Enable free shipping
	//
	// Adds and returns a FREE shipping option (can be used for digital products)
	// in:body
	// example: true
	EnableFreeShipping bool `json:"enable_free_shipping"`
}

// swagger:parameters cart GetTaxRateRequest
type GetTaxRateRequest struct {
	// Body of the request
	//
	// required: true
	// in:body
	Body *GetTaxRateRequestBody
}

type GetTaxRateRequestBody struct {
	// Shipping Address ID
	//
	// required: true
	// in:body
	// example: 123e4567-e89b-12d3-a456-426614174000
	AddressID uuid.UUID `json:"address_id"`
	// Shipping Rate ID selected by the customer
	//
	// required: true
	// in:body
	// example: "123e4567-e89b-12d3-a456-426614174000"
	ShippingRateID *uuid.UUID `json:"shipping_rate_id"`
	// Warehouse address
	//
	// required: true
	// in:body
	// example:
	WarehouseAddress *WarehouseAddress `json:"warehouse_address"`
}

type WarehouseAddress struct {
	City        string `json:"city"`
	StateCode   string `json:"state_code"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

type CreateCartShippingRatesRequest struct {
	// Body of the request
	//
	// required: true
	// in:body
	Body *CreateCartShippingRatesRequestBody
}

type CreateCartShippingRatesRequestBody struct {
	// Address ID
	//
	// required: true
	// in:body
	// example: "123e4567-e89b-12d3-a456-426614174000"
	AddressID uuid.UUID `json:"address_id"`
	// Cart Shipping Rates
	//
	// required: true
	// in:body
	// example: [{"amount": "100.00", "currency": "USD", "carrier_name": "UPS", "carrier_code": "UPS", "service_type": "Standard", "service_code": "123456", "estimated_delivery_date": "2021-01-01", "business_days_in_transit": "3"}]
	CartShippingRates []CartShippingRateRequest `json:"cart_shipping_rates"`
}

type CartShippingRateRequest struct {
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	CarrierName           string          `json:"carrier_name"`
	CarrierCode           string          `json:"carrier_code"`
	ServiceType           string          `json:"service_type"`
	ServiceCode           string          `json:"service_code"`
	EstimatedDeliveryDate time.Time       `json:"estimated_delivery_date"`
	BusinessDaysInTransit string          `json:"business_days_in_transit"`
}

// swagger:parameters cart SetCartItemShippingRateRequest
type SetCartItemShippingRateRequest struct {
	// Body of the request
	//
	// required: true
	// in:body
	Body *SetCartItemShippingRateRequestBody
}

type SetCartItemShippingRateRequestBody struct {
	// Cart Item ID
	//
	// required: true
	// in:body
	// example: "123e4567-e89b-12d3-a456-426614174000"
	CartItemID uuid.UUID `json:"cart_item_id"`
	// Shipping Rate ID selected for this cart item
	//
	// required: true
	// in:body
	// example: "456e7890-e89b-12d3-a456-426614174001"
	ShippingRateID uuid.UUID `json:"shipping_rate_id"`
}
