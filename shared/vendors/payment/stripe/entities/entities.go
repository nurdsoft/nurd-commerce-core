package entities

import "github.com/shopspring/decimal"

type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type CreateCustomerResponse struct {
	Id string `json:"id"`
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

type GetCustomerPaymentMethodsResponse struct {
	PaymentMethods []PaymentMethod `json:"payment_methods"`
}

type GetSetupIntentResponse struct {
	SetupIntent  string
	EphemeralKey string
	CustomerId   string
}

type CreatePaymentIntentRequest struct {
	Amount          decimal.Decimal
	Currency        string
	CustomerId      *string
	PaymentMethodId string
}

type CreatePaymentIntentResponse struct {
	Id string `json:"id"`
}

type HandleWebhookEventRequest struct {
	Payload   []byte
	Signature string
}

type HandleWebhookEventResponse struct {
	ObjectId string
	Type     string
}

type GetCustomerPaymentMethodResponse struct {
	PaymentMethod PaymentMethod `json:"payment_methods"`
}

type RefundRequest struct {
	PaymentIntentId string          `json:"payment_intent_id"`
	Amount          decimal.Decimal `json:"amount"`
}

type RefundResponse struct {
	// Id of the refund
	Id string
	// Amount refunded
	Amount decimal.Decimal
	// Currency of the refund
	Currency string
	// Status of the refund. This can be `pending`, `requires_action`, `succeeded`, `failed`, or `canceled`.
	// Learn more about [failed refunds](https://stripe.com/docs/refunds#failed-refunds).
	Status string
	Reason string
	// TODO Add destination details if needed
}
