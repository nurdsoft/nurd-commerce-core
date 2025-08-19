package entities

import "github.com/shopspring/decimal"

// swagger:model CreateOrderResponse
type CreateOrderResponse struct {
	// Order reference
	//
	// example: VZQ9IMMMYQ
	OrderReference string `json:"order_reference"`
	// Order items
	OrderItems []*OrderItem `json:"order_items"`
}

// swagger:model ListOrdersResponse
type ListOrdersResponse struct {
	Orders     []*Order `json:"orders"`
	NextCursor string   `json:"next_cursor"`
}

type GetOrderData struct {
	Order      *Order       `json:"order"`
	OrderItems []*OrderItem `json:"order_items"`
}

// swagger:response GetOrderResponse
type GetOrderResponse struct {
	// in: body
	Body struct {
		Data GetOrderData `json:"data"`
	}
}

// swagger:model RefundOrderResponse
type RefundOrderResponse struct {
	// Total amount that will be refunded
	TotalRefundableAmount decimal.Decimal `json:"total_refundable_amount"`
	// Items that will be refunded
	RefundableItems []*RefundableItem `json:"refundable_items"`
}

type RefundableItem struct {
	// Order Item ID
	ItemId string `json:"item_id"`
	// Order Item SKU
	Sku string `json:"sku"`
	// Order Item Quantity
	Quantity int `json:"quantity"`
	// Order Item Price that will be refunded
	Price decimal.Decimal `json:"price"`
	// Refund Initiated
	RefundInitiated bool `json:"refund_initiated"`
}
