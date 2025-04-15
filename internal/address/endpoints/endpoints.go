package endpoints

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/service"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	AddAddressEndpoint      endpoint.Endpoint
	GetAllAddressesEndpoint endpoint.Endpoint
	GetAddressEndpoint      endpoint.Endpoint
	UpdateAddressEndpoint   endpoint.Endpoint
	DeleteAddressEndpoint   endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		AddAddressEndpoint:      makeAddAddress(svc),
		GetAllAddressesEndpoint: makeGetAllAddresses(svc),
		GetAddressEndpoint:      makeGetAddress(svc),
		UpdateAddressEndpoint:   makeUpdateAddress(svc),
		DeleteAddressEndpoint:   makeDeleteAddress(svc),
	}
}

func makeAddAddress(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.AddAddressRequest) //nolint:errcheck
		return svc.AddAddress(ctx, req)
	}
}

func makeGetAllAddresses(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return svc.GetAddresses(ctx)
	}
}

func makeGetAddress(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetAddressRequest) //nolint:errcheck

		return svc.GetAddress(ctx, req)
	}
}

func makeUpdateAddress(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.UpdateAddressRequest) //nolint:errcheck
		return svc.UpdateAddress(ctx, req)
	}
}

func makeDeleteAddress(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.DeleteAddressRequest) //nolint:errcheck
		return nil, svc.DeleteAddress(ctx, req)
	}
}
