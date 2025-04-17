package client

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/service"
)

type Client interface {
	GetAccountByID(ctx context.Context, accountId string) (*entities.Account, error)
	CreateUserAccount(ctx context.Context, req *entities.CreateSFUserRequest) (*entities.CreateSFUserResponse, error)
	UpdateUserAccount(ctx context.Context, req *entities.UpdateSFUserRequest) error
	CreateUserAddress(ctx context.Context, req *entities.CreateSFAddressRequest) (*entities.CreateSFAddressResponse, error)
	UpdateUserAddress(ctx context.Context, req *entities.UpdateSFAddressRequest) error
	DeleteUserAddress(ctx context.Context, addressId string) error
	CreateProduct(ctx context.Context, req *entities.CreateSFProductRequest) (*entities.CreateSFProductResponse, error)
	CreatePriceBookEntry(ctx context.Context, req *entities.CreateSFPriceBookEntryRequest) (*entities.CreateSFPriceBookEntryResponse, error)
	CreateOrder(ctx context.Context, req *entities.CreateSFOrderRequest) (*entities.CreateSFOrderResponse, error)
	AddOrderItems(ctx context.Context, items []*entities.OrderItem) (*entities.AddOrderItemResponse, error)
	UpdateOrderStatus(ctx context.Context, req *entities.UpdateOrderRequest) error
	GetOrderItems(ctx context.Context, orderId string) (*entities.GetOrderItemsResponse, error)
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (l localClient) GetAccountByID(ctx context.Context, accountId string) (*entities.Account, error) {
	return l.svc.GetAccountByID(ctx, accountId)
}

func (l localClient) CreateUserAccount(ctx context.Context, req *entities.CreateSFUserRequest) (*entities.CreateSFUserResponse, error) {
	return l.svc.CreateUserAccount(ctx, req)
}

func (l localClient) UpdateUserAccount(ctx context.Context, req *entities.UpdateSFUserRequest) error {
	return l.svc.UpdateUserAccount(ctx, req)
}

func (l localClient) CreateUserAddress(ctx context.Context, req *entities.CreateSFAddressRequest) (*entities.CreateSFAddressResponse, error) {
	return l.svc.CreateUserAddress(ctx, req)
}

func (l localClient) UpdateUserAddress(ctx context.Context, req *entities.UpdateSFAddressRequest) error {
	return l.svc.UpdateUserAddress(ctx, req)
}

func (l localClient) DeleteUserAddress(ctx context.Context, addressId string) error {
	return l.svc.DeleteUserAddress(ctx, addressId)
}

func (l localClient) CreateProduct(ctx context.Context, req *entities.CreateSFProductRequest) (*entities.CreateSFProductResponse, error) {
	return l.svc.CreateProduct(ctx, req)
}

func (l localClient) CreatePriceBookEntry(ctx context.Context, req *entities.CreateSFPriceBookEntryRequest) (*entities.CreateSFPriceBookEntryResponse, error) {
	return l.svc.CreatePriceBookEntry(ctx, req)
}

func (l localClient) CreateOrder(ctx context.Context, req *entities.CreateSFOrderRequest) (*entities.CreateSFOrderResponse, error) {
	return l.svc.CreateOrder(ctx, req)
}

func (l localClient) AddOrderItems(ctx context.Context, items []*entities.OrderItem) (*entities.AddOrderItemResponse, error) {
	return l.svc.AddOrderItems(ctx, items)
}

func (l localClient) UpdateOrderStatus(ctx context.Context, req *entities.UpdateOrderRequest) error {
	return l.svc.UpdateOrderStatus(ctx, req)
}

func (l localClient) GetOrderItems(ctx context.Context, orderId string) (*entities.GetOrderItemsResponse, error) {
	return l.svc.GetOrderItems(ctx, orderId)
}
