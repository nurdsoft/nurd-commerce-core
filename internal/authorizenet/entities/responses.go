package entities

// swagger:model GetPaymentProfileResponse
// GetPaymentProfileResponse is the response for getting payment profiles
type GetPaymentProfileResponse struct {
	PaymentProfiles []PaymentProfile `json:"payment_profiles"`
}

type PaymentProfile struct {
	ID             string `json:"id"`
	CardNumber     string `json:"card_number"`
	CardType       string `json:"card_type"`
	ExpirationDate string `json:"expiration_date"`
}

// swagger:model CreatePaymentProfileResponse
type CreatePaymentProfileResponse struct {
	ProfileID        string `json:"profile_id"`
	PaymentProfileID string `json:"payment_profile_id"`
}
