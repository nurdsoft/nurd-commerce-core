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

// swagger:parameters stripe GetPaymentMethodsRequest
type StripeGetPaymentMethodRequest struct {
	// Payment Method ID
	//
	// required: true
	// in:path
	// example: pm_1J2Y3Z4A5B6C7D8E9F0G
	PaymentMethodId string `json:"payment_method_id"`
}

// swagger:parameters stripe StripeRefundRequest
type StripeRefundRequest struct {
	// Payment Intent ID
	//
	// required: true
	// in:path
	// example: pi_1J2Y3Z4A5B6C7D8E9F0G
	PaymentIntentId string `json:"payment_intent_id"`
	// Body
	//
	// in:body
	Body *StripeRefundRequestBody
}

type StripeRefundRequestBody struct {
	Amount string `json:"amount"`
}
