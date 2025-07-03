package payment

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
)

type Client interface {
	CreatePayment(ctx context.Context, req any) (providers.PaymentProviderResponse, error)
	GetProvider() providers.ProviderType
	Refund(ctx context.Context, req any) (*providers.RefundResponse, error)
}
