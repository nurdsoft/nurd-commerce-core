package client

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/service"
	"github.com/pkg/errors"
)

type Client interface {
	CreateCustomer(ctx context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error)
	GetCustomerPaymentMethods(ctx context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error)
	GetCustomerPaymentMethodById(_ context.Context, customerId, paymentMethodId *string) (*entities.GetCustomerPaymentMethodResponse, error)
	GetSetupIntent(ctx context.Context, customerId *string) (*entities.GetSetupIntentResponse, error)
	CreatePayment(ctx context.Context, req any) (providers.PaymentProviderResponse, error)
	GetWebhookEvent(ctx context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error)
	GetProvider() providers.ProviderType
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) CreateCustomer(ctx context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error) {
	return c.svc.CreateCustomer(ctx, req)
}

func (c *localClient) GetCustomerPaymentMethods(ctx context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error) {
	return c.svc.GetCustomerPaymentMethods(ctx, customerId)
}

func (c *localClient) GetCustomerPaymentMethodById(ctx context.Context, customerId, paymentMethodId *string) (*entities.GetCustomerPaymentMethodResponse, error) {
	return c.svc.GetCustomerPaymentMethodById(ctx, customerId, paymentMethodId)
}

func (c *localClient) GetSetupIntent(ctx context.Context, customerId *string) (*entities.GetSetupIntentResponse, error) {
	return c.svc.GetSetupIntent(ctx, customerId)
}

func (c *localClient) GetWebhookEvent(ctx context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error) {
	return c.svc.GetWebhookEvent(ctx, req)
}

func (c *localClient) CreatePayment(ctx context.Context, req any) (providers.PaymentProviderResponse, error) {
	stripeReq, ok := req.(entities.CreatePaymentIntentRequest)
	if !ok {
		return providers.PaymentProviderResponse{}, errors.New("invalid request type")
	}

	res, err := c.svc.CreatePaymentIntent(ctx, &stripeReq)
	if err != nil {
		return providers.PaymentProviderResponse{}, err
	}

	return providers.PaymentProviderResponse{
		ID:     res.Id,
		Status: providers.PaymentStatusPending,
	}, nil
}

func (c *localClient) GetProvider() providers.ProviderType {
	return providers.ProviderStripe
}
