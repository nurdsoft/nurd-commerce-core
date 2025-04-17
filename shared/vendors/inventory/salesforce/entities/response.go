package entities

type CreateSFUserResponse struct {
	ID      string `json:"Id"`
	Success bool   `json:"Success"`
}

type CreateSFAddressResponse struct {
	ID      string `json:"Id"`
	Success bool   `json:"Success"`
}

type CreateSFProductResponse struct {
	ID      string `json:"Id"`
	Success bool   `json:"Success"`
}

type CreateSFPriceBookEntryResponse struct {
	ID      string `json:"Id"`
	Success bool   `json:"Success"`
}

type CreateSFOrderResponse struct {
	ID      string `json:"Id"`
	Success bool   `json:"Success"`
}
type AddOrderItemResponse struct {
	HasErrors bool      `json:"hasErrors"`
	Results   []Results `json:"results"`
}
type Result struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Errors  []any  `json:"errors"`
}
type Results struct {
	StatusCode int    `json:"statusCode"`
	Result     Result `json:"result"`
}

type GetOrderItemsResponse struct {
	TotalSize int       `json:"totalSize"`
	Done      bool      `json:"done"`
	Records   []Records `json:"records"`
}
type Attributes struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}
type Records struct {
	// Attributes       Attributes `json:"attributes"`
	ID               string  `json:"Id"`
	Quantity         float64 `json:"Quantity"`
	Product2ID       string  `json:"Product2Id"`
	Description      string  `json:"Description"`
	PricebookEntryID string  `json:"PricebookEntryId"`
	TypeC            string  `json:"Type__c"`
}
