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
