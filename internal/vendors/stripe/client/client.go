package client

import (
	"context"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/service"
)

type Client interface {
	CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error)
	CreateCustomer(ctx context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error)
	GetCustomerPaymentMethods(ctx context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error)
	GetSetupIntent(ctx context.Context, customerId *string) (*entities.GetSetupIntentResponse, error)
	CreatePaymentIntent(ctx context.Context, req *entities.CreatePaymentIntentRequest) (*entities.CreatePaymentIntentResponse, error)
	GetWebhookEvent(ctx context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error)
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error) {
	return c.svc.CalculateTax(ctx, req)
}

func (c *localClient) CreateCustomer(ctx context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error) {
	return c.svc.CreateCustomer(ctx, req)
}

func (c *localClient) GetCustomerPaymentMethods(ctx context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error) {
	return c.svc.GetCustomerPaymentMethods(ctx, customerId)
}

func (c *localClient) GetSetupIntent(ctx context.Context, customerId *string) (*entities.GetSetupIntentResponse, error) {
	return c.svc.GetSetupIntent(ctx, customerId)
}

func (c *localClient) CreatePaymentIntent(ctx context.Context, req *entities.CreatePaymentIntentRequest) (*entities.CreatePaymentIntentResponse, error) {
	return c.svc.CreatePaymentIntent(ctx, req)
}

func (c *localClient) GetWebhookEvent(ctx context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error) {
	return c.svc.GetWebhookEvent(ctx, req)
}
