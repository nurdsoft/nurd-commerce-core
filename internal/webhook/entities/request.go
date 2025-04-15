package entities


type NotifyOrderStatusChangeRequest struct {
	OrderID        string `json:"order_id"`
	OrderReference string `json:"order_reference"`
	Status         string `json:"status"`
}
