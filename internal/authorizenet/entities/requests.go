package entities

// swagger:model WebhookRequest
type WebhookRequest struct {
	NotificationID string  `json:"notificationId"`
	EventType      string  `json:"eventType"`
	EventDate      string  `json:"eventDate"`
	WebhookID      string  `json:"webhookId"`
	Payload        Payload `json:"payload"`
}

type Payload struct {
	ResponseCode int         `json:"responseCode"`
	AvsResponse  string      `json:"avsResponse"`
	AuthAmount   float64     `json:"authAmount"`
	FraudList    []FraudItem `json:"fraudList"`
	EntityName   string      `json:"entityName"`
	ID           string      `json:"id"`
}

type FraudItem struct {
	FraudFilter string `json:"fraudFilter"`
	FraudAction string `json:"fraudAction"`
}

// swagger:model CreatePaymentProfileRequest
type CreatePaymentProfileRequest struct {
	CardNumber     string `json:"card_number"`
	ExpirationDate string `json:"expiration_date"`
}

// swagger:parameters authorizenet CreatePaymentProfileRequest
// CreatePaymentProfileRequestBody wraps the request body for creating a payment profile
// required: true
// in:body
type CreatePaymentProfileRequestBody struct {
	// The payment profile to create
	// required: true
	// in:body
	Body CreatePaymentProfileRequest `json:"body"`
}

// swagger:parameters authorizenet WebhookRequest
// WebhookRequestBody wraps the request body for webhook events
// required: true
// in:body
type WebhookRequestBody struct {
	// The webhook event data
	// required: true
	// in:body
	Body WebhookRequest `json:"body"`
}
