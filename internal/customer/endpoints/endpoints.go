package endpoints

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/service"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	CreateCustomerEndpoint endpoint.Endpoint
	UpdateCustomerEndpoint endpoint.Endpoint
	GetCustomerEndpoint    endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		CreateCustomerEndpoint: makeCreateCustomer(svc),
		UpdateCustomerEndpoint: makeUpdateCustomer(svc),
		GetCustomerEndpoint:    makeGetCustomer(svc),
	}
}

func makeCreateCustomer(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.CreateCustomerRequest) //nolint:errcheck

		return svc.CreateCustomer(ctx, req)
	}
}

func makeUpdateCustomer(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.UpdateCustomerRequest) //nolint:errcheck

		return svc.UpdateCustomer(ctx, req)
	}
}

func makeGetCustomer(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return svc.GetCustomer(ctx)
	}
}
