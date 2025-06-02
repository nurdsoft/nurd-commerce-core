package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/nurdsoft/nurd-commerce-core/internal/authorizenet/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/authorizenet/service"
)

type Endpoints struct {
	GetPaymentProfilesEndpoint   endpoint.Endpoint
	CreatePaymentProfileEndpoint endpoint.Endpoint
	WebhookEndpoint              endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		GetPaymentProfilesEndpoint:   makeGetPaymentProfiles(svc),
		CreatePaymentProfileEndpoint: makeCreatePaymentProfile(svc),
		WebhookEndpoint:              makeWebhookEndpoint(svc),
	}
}

func makeGetPaymentProfiles(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return svc.GetPaymentProfiles(ctx)
	}
}

func makeCreatePaymentProfile(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(entities.CreatePaymentProfileRequestBody)
		return svc.CreatePaymentProfile(ctx, req)
	}
}

func makeWebhookEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(entities.WebhookRequestBody)
		err := svc.HandleWebhook(ctx, req)
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success"}, nil
	}
}
