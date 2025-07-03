package providers

type (
	ProviderType  string
	PaymentStatus string
)

const (
	ProviderStripe       ProviderType = "stripe"
	ProviderAuthorizeNet ProviderType = "authorizeNet"
)

const (
	PaymentStatusSuccess PaymentStatus = "success"
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusFailed  PaymentStatus = "failed"
)

type PaymentProviderResponse struct {
	ID     string
	Status PaymentStatus
}

type RefundResponse struct {
	ID     string
	Status string
}
