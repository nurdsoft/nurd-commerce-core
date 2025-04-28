package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

// swagger:parameters orders CreateOrderRequest
type CreateOrderRequest struct {
	// Body of the request
	//
	// in:body
	Body *CreateOrderRequestBody
}

type CreateOrderRequestBody struct {
	AddressID             uuid.UUID `json:"address_id"`
	ShippingRateID        uuid.UUID `json:"shipping_rate_id"`
	StripePaymentMethodID string    `json:"stripe_payment_method_id,omitempty"`
}

// swagger:parameters orders ListOrdersRequest
type ListOrdersRequest struct {
	// Limit of orders to return
	//
	// required: true
	// in:query
	Limit int `json:"limit"`
	// Cursor to paginate orders
	//
	// in:query
	Cursor string `json:"cursor"`
}

// swagger:parameters orders GetOrderRequest
type GetOrderRequest struct {
	// Order ID
	//
	// required:true
	// in:path
	OrderID uuid.UUID `json:"order_id"`
}

// swagger:parameters orders CancelOrderRequest
type CancelOrderRequest struct {
	// Order ID
	//
	// required:true
	// in:path
	OrderID uuid.UUID `json:"order_id"`
}

// swagger:parameters orders UpdateOrderRequest
type UpdateOrderRequest struct {
	// Order reference
	//
	// required:true
	// in:path
	OrderReference string `json:"order_reference"`
	// Body of the request
	//
	// in:body
	Body *UpdateOrderRequestBody
}

type UpdateOrderRequestBody struct {
	Status                     *string          `json:"status"`
	FulfillmentMessage         *string          `json:"fulfillment_message,omitempty"`
	FulfillmentShipmentDate    *time.Time       `json:"fulfillment_shipment_date,omitempty"`
	FulfillmentFreightCharge   *decimal.Decimal `json:"fulfillment_freight_charge,omitempty"`
	FulfillmentOrderTotal      *decimal.Decimal `json:"fulfillment_order_total,omitempty"`
	FulfillmentVendorAmountDue *decimal.Decimal `json:"fulfillment_vendor_amount_due,omitempty"`
}
