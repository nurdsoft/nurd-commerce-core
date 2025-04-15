package entities

// swagger:model CreateOrderResponse
type CreateOrderResponse struct {
	// Order reference
	//
	// example: VZQ9IMMMYQ
	OrderReference string `json:"order_reference"`
}

// swagger:model ListOrdersResponse
type ListOrdersResponse struct {
	Orders     []*Order `json:"orders"`
	NextCursor string   `json:"next_cursor"`
}

// swagger:model GetOrderResponse
type GetOrderResponse struct {
	Order      *Order       `json:"order"`
	OrderItems []*OrderItem `json:"order_items"`
}
