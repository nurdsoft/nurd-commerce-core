package entities

// swagger:parameters webhook StripeWebhookRequest
type StripeWebhookRequest struct {
	// Payload
	//
	// required: true
	// in:body
	Payload []byte `json:"payload"`
	// Signature
	//
	// required: true
	// in:body
	Signature string `json:"signature"`
}
