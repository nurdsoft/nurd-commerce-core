package endpoints

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/entities"

	"github.com/go-kit/kit/endpoint"
	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/service"
)

type Endpoints struct {
	StripeGetPaymentMethodsEndpoint endpoint.Endpoint
	StripeGetPaymentMethodEndpoint  endpoint.Endpoint
	StripeGetSetupIntentEndpoint    endpoint.Endpoint
	StripeWebhookEndpoint           endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		StripeGetPaymentMethodsEndpoint: makeStripeGetPaymentMethods(svc),
		StripeGetPaymentMethodEndpoint:  makeStripeGetPaymentMethod(svc),
		StripeGetSetupIntentEndpoint:    makeStripeGetSetupIntent(svc),
		StripeWebhookEndpoint:           makeStripeWebhookEndpoint(svc),
	}
}

func makeStripeGetPaymentMethods(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return svc.GetPaymentMethods(ctx)
	}
}

func makeStripeGetPaymentMethod(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.StripeGetPaymentMethodRequest)
		return svc.GetPaymentMethod(ctx, req)
	}
}

func makeStripeGetSetupIntent(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return svc.GetSetupIntent(ctx)
	}
}

func makeStripeWebhookEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.StripeWebhookRequest)
		err := svc.HandleStripeWebhook(ctx, req)
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success"}, nil
	}
}
