package entities

import "github.com/shopspring/decimal"

type CreateCustomerRequest struct {
	CustomerID  string
	Description string
	Email       string
}

type CreateCustomerResponse struct {
	ProfileID string
}

type CreateCustomerPaymentProfileRequest struct {
	ProfileID      string
	CardNumber     string
	ExpirationDate string
}

type CreateCustomerPaymentProfileResponse struct {
	ProfileID        string
	PaymentProfileID string
}

type GetPaymentProfilesRequest struct {
	ProfileID string
}

type GetPaymentProfilesResponse struct {
	PaymentProfiles []PaymentProfile
}

type PaymentProfile struct {
	ID             string
	CardNumber     string
	CardType       string
	ExpirationDate string
}

type CreatePaymentTransactionRequest struct {
	Amount       decimal.Decimal
	PaymentNonce string
	BillingInfo  BillingInfo
}

type BillingInfo struct {
	FirstName string
	LastName  string
	Address   string
	City      string
	State     string
	Country   string
	Zip       string
}

type CreatePaymentTransactionResponse struct {
	ID     string
	Status string
}

type HandleWebhookEventRequest struct{}
type HandleWebhookEventResponse struct{}
