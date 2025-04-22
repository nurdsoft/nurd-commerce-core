package endpoints

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/service"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	AddToWishlistEndpoint       endpoint.Endpoint
	RemoveFromWishlistEndpoint  endpoint.Endpoint
	GetWishlistEndpoint         endpoint.Endpoint
	GetMoreFromWishlistEndpoint endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		AddToWishlistEndpoint:       makeAddToWishlist(svc),
		RemoveFromWishlistEndpoint:  makeRemoveFromWishlist(svc),
		GetWishlistEndpoint:         makeGetWishlist(svc),
		GetMoreFromWishlistEndpoint: makeGetMoreFromWishlist(svc),
	}
}

func makeAddToWishlist(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.AddToWishlistRequest) //nolint:errcheck
		err := svc.AddToWishlist(ctx, req)

		return nil, err
	}
}

func makeRemoveFromWishlist(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.RemoveFromWishlistRequest) //nolint:errcheck
		err := svc.RemoveFromWishlist(ctx, req)

		return nil, err
	}
}

func makeGetWishlist(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetWishlistRequest) //nolint:errcheck
		wishlist, err := svc.GetWishlist(ctx, req)

		return wishlist, err
	}
}

func makeGetMoreFromWishlist(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetMoreFromWishlistRequest) //nolint:errcheck
		return svc.GetMoreFromWishlist(ctx, req)
	}
}
