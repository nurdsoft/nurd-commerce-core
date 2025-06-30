package entities

// swagger:model GetPaymentMethodsResponse
type GetPaymentMethodsResponse struct {
	PaymentMethods []PaymentMethod `json:"payment_methods"`
}

// swagger:model GetSetupIntentResponse
type GetSetupIntentResponse struct {
	SetupIntent SetupIntent `json:"setup_intent"`
}

type PaymentMethod struct {
	Id           string  `json:"id"`
	Brand        string  `json:"brand"`
	DisplayBrand string  `json:"display_brand"`
	Country      string  `json:"country"`
	Last4        string  `json:"last4"`
	ExpiryMonth  int64   `json:"expiry_month"`
	ExpiryYear   int64   `json:"expiry_year"`
	Wallet       *string `json:"wallet"`
	Created      int64   `json:"created"`
}

type SetupIntent struct {
	// Client secret of the SetupIntent.
	SetupIntent string `json:"setup_intent"`
	// Short-lived key associated with a specific Stripe Customer, will be used by StripeSDK
	EphemeralKey string `json:"ephemeral_key"`
	// Customer ID associated with the SetupIntent
	CustomerId string `json:"customer"`
}

// swagger:model GetPaymentMethodResponse
type GetPaymentMethodResponse struct {
	PaymentMethod PaymentMethod `json:"payment_method"`
}

// swagger:model StripeRefundResponse
type StripeRefundResponse struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}
