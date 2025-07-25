package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"
)

// swagger:parameters orders CreateOrderRequest
type CreateOrderRequest struct {
	// Body of the request
	//
	// in:body
	Body *CreateOrderRequestBody
}

type CreateOrderRequestBody struct {
	AddressID             uuid.UUID   `json:"address_id"`
	ShippingRateID        *uuid.UUID   `json:"shipping_rate_id"`
	StripePaymentMethodID string      `json:"stripe_payment_method_id,omitempty"`
	PaymentNonce          string      `json:"payment_nonce,omitempty"`
	BillingInfo           BillingInfo `json:"billing_info,omitzero"`
}

type BillingInfo struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Address   string `json:"address,omitempty"`
	City      string `json:"city,omitempty"`
	State     string `json:"state,omitempty"`
	Country   string `json:"country,omitempty"`
	Zip       string `json:"zip,omitempty"`
}

type CreatePaymentRequest struct {
	Amount          decimal.Decimal
	Currency        string
	Customer        entities.Customer
	PaymentMethodId string
	PaymentNonce    string
	BillingInfo     BillingInfo
}

// swagger:parameters orders ListOrdersRequest
type ListOrdersRequest struct {
	// Limit of orders to return
	//
	// required: true
	// in:query
	// example: 10
	Limit int `json:"limit"`
	// Cursor to paginate orders
	//
	// in:query
	// example: MjAyNS0wNS0yNlQxNjo0Mjo1MSswNTozMA==
	Cursor string `json:"cursor"`
	// Boolean to indicate whether order items should be included in the response.
	// If true, the response will include an additional `items_summary` field in the response for each order item.
	//
	// in:query
	IncludeItems bool `json:"include_items,omitempty"`
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
	Status                    *string          `json:"status"`
	FulfillmentMessage        *string          `json:"fulfillment_message,omitempty"`
	FulfillmentShipmentDate   *time.Time       `json:"fulfillment_shipment_date,omitempty"`
	FulfillmentFreightCharge  *decimal.Decimal `json:"fulfillment_freight_charge,omitempty"`
	FulfillmentOrderTotal     *decimal.Decimal `json:"fulfillment_order_total,omitempty"`
	FulfillmentAmountDue      *decimal.Decimal `json:"fulfillment_amount_due,omitempty"`
	FulfillmentTrackingNumber *string          `json:"fulfillment_tracking_number,omitempty"`
	FulfillmentTrackingURL    *string          `json:"fulfillment_tracking_url,omitempty"`
	FulfilmentMetadata        *json.JSON       `json:"fulfillment_metadata,omitempty"`
	Items                     []*Item          `json:"items,omitempty"`
}

type Item struct {
	ID     uuid.UUID        `json:"id"`
	Sku    string           `json:"sku"`
	Status *OrderItemStatus `json:"status,omitempty"`
}

// swagger:parameters orders RefundOrderRequest
type RefundOrderRequest struct {
	// Order reference
	//
	// required:true
	// in:path
	OrderReference string `json:"order_reference"`
	// Body of the request
	//
	// in:body
	Body *RefundOrderRequestBody
}

type RefundOrderRequestBody struct {
	Items []*RefundItem `json:"items"`
}

type RefundItem struct {
	// Unique identifier for the item to refund
	Sku string `json:"sku"`
	// Quantity of the item to refund
	Quantity int `json:"quantity"`
}
