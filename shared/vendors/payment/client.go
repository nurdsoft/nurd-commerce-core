package payment

import (
	"context"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
)

type Client interface {
	CreateCustomer(_ context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error)
	GetCustomerPaymentMethods(_ context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error)
	GetSetupIntent(_ context.Context, customerId *string) (*entities.GetSetupIntentResponse, error)
	CreatePaymentIntent(ctx context.Context, req *entities.CreatePaymentIntentRequest) (*entities.CreatePaymentIntentResponse, error)
	GetWebhookEvent(_ context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error)
}
