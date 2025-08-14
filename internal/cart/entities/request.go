package entities

import (
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
	// example: 123e4567-e89b-12d3-a456-426614174000
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
