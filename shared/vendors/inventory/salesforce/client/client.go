package client

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	ordersRepo "github.com/nurdsoft/nurd-commerce-core/internal/orders/repository"
	productRepo "github.com/nurdsoft/nurd-commerce-core/internal/product/repository"
	inventoryEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/providers"
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
	CreateOrder(ctx context.Context, req inventoryEntities.CreateInventoryOrderRequest) (any, error)
	AddOrderItems(ctx context.Context, items []*entities.OrderItem) (*entities.AddOrderItemResponse, error)
	UpdateOrderStatus(ctx context.Context, req inventoryEntities.UpdateInventoryOrderStatusRequest) error
	GetOrderItems(ctx context.Context, orderId string) (*entities.GetOrderItemsResponse, error)
	GetProductByID(ctx context.Context, id string) (*inventoryEntities.Product, error)
}

func NewClient(svc service.Service, provider providers.ProviderType, productsRepo productRepo.Repository, ordersRepo ordersRepo.Repository) Client {
	return &localClient{svc, productsRepo, ordersRepo}
}

type localClient struct {
	svc          service.Service
	productsRepo productRepo.Repository
	ordersRepo   ordersRepo.Repository
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

func (l localClient) CreateOrder(ctx context.Context, req inventoryEntities.CreateInventoryOrderRequest) (any, error) {
	if req.Customer.SalesforceID == nil {
		return nil, errors.New("customer salesforce id is required")
	}

	city := ""
	if req.Address.City != nil {
		city = *req.Address.City
	}

	// create order on salesforce
	res, err := l.svc.CreateOrder(ctx, &entities.CreateSFOrderRequest{
		OrderReferenceC:         req.Order.OrderReference,
		AccountID:               *req.Customer.SalesforceID,
		EffectiveDate:           time.Now().Format("2006-01-02"),
		Status:                  req.Order.Status.String(),
		BillingStreet:           req.Address.Address,
		BillingCity:             city,
		BillingState:            req.Address.StateCode,
		BillingPostalCode:       req.Address.PostalCode,
		BillingCountry:          req.Address.CountryCode,
		ShippingStreet:          req.Address.Address,
		ShippingCity:            city,
		ShippingState:           req.Address.StateCode,
		ShippingPostalCode:      req.Address.PostalCode,
		ShippingCountry:         req.Address.CountryCode,
		TotalC:                  req.Order.Total.String(),
		SubTotalC:               req.Order.Subtotal.String(),
		ShippingRateC:           req.Order.ShippingRate.String(),
		TaxAmountC:              req.Order.TaxAmount.String(),
		ShippingCarrierNameC:    req.Order.ShippingCarrierName,
		ShippingCarrierServiceC: req.Order.ShippingServiceType,
		CurrencyC:               req.Order.Currency,
		Pricebook2ID:            entities.StandardPriceBook,
		EstimatedDeliveryDateC: func() string {
			if req.Order.ShippingEstimatedDeliveryDate.IsZero() {
				return time.Now().Format("2006-01-02")
			}
			return req.Order.ShippingEstimatedDeliveryDate.Format("2006-01-02")
		}(),
		OrderCreatedAtC: time.Now().Format("2006-01-02"),
	})
	if err != nil {
		return nil, err
	}
	if res.Success {
		// update order with salesforce order id
		err = l.ordersRepo.Update(ctx, map[string]interface{}{
			"salesforce_id": res.ID,
		}, req.Order.ID.String(), req.Order.CustomerID.String())
		if err != nil {
			return nil, err
		}

		productIDs := func() []string {
			var ids []string
			for _, item := range req.CartItems {
				ids = append(ids, item.ProductID.String())
			}
			return ids
		}
		// get salesforce products by product ids
		products, err := l.productsRepo.FindByIDs(ctx, productIDs())
		if err != nil {
			return nil, err
		}

		sfOrderItems := make([]*entities.OrderItem, 0, len(req.CartItems))

		if len(products) > 0 {
			for _, item := range req.CartItems {

				description := ""
				if item.Description != nil {
					description = *item.Description
				}

				sfOrderItem := entities.OrderItem{
					OrderID:     res.ID,
					Quantity:    item.Quantity,
					UnitPrice:   item.Price.InexactFloat64(),
					Description: item.Name,
					TypeC:       description,
				}

				for _, product := range products {
					if product.ID == item.ProductID && product.SalesforcePricebookEntryId != nil {
						sfOrderItem.PricebookEntryID = *product.SalesforcePricebookEntryId
						break
					}
				}

				sfOrderItems = append(sfOrderItems, &sfOrderItem)
			}

			// add items to order on salesforce
			_, err = l.svc.AddOrderItems(ctx, sfOrderItems)
			if err != nil {
				return nil, err
			}

			// get order items from salesforce
			items, err := l.svc.GetOrderItems(ctx, res.ID)
			if err != nil {
				return nil, err
			}

			if len(items.Records) > 0 {
				// map of order item id to salesforce order item id
				ids := make(map[string]string)
				// build the map based on TypeC and Description and ProductID & Product2Id
				// salesforce order items
				for _, item := range items.Records {
					// order items
					for _, orderItem := range req.OrderItems {
						if orderItem.Description != nil &&
							item.TypeC == *orderItem.Description &&
							item.Description == orderItem.Name {
							ids[orderItem.ID.String()] = item.ID
							break
						}
					}
				}

				// update order items with salesforce order item ids
				err = l.ordersRepo.AddSalesforceIDPerOrderItem(ctx, ids)
				if err != nil {
					return nil, err
				}
			}
		} else {
			return nil, errors.New("no products found for order items")
		}
	}

	return res, nil
}

func (l localClient) AddOrderItems(ctx context.Context, items []*entities.OrderItem) (*entities.AddOrderItemResponse, error) {
	return l.svc.AddOrderItems(ctx, items)
}

func (l localClient) UpdateOrderStatus(ctx context.Context, req inventoryEntities.UpdateInventoryOrderStatusRequest) error {
	if req.Order.SalesforceID != "" && req.Customer.SalesforceID != nil {
		return l.svc.UpdateOrderStatus(ctx, &entities.UpdateOrderRequest{
			OrderId:   req.Order.SalesforceID,
			AccountID: *req.Customer.SalesforceID,
			Status:    req.Status,
		})
	}

	return nil
}

func (l localClient) GetOrderItems(ctx context.Context, orderId string) (*entities.GetOrderItemsResponse, error) {
	return l.svc.GetOrderItems(ctx, orderId)
}

func (l localClient) GetProductByID(ctx context.Context, id string) (*inventoryEntities.Product, error) {
	product, err := l.productsRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &inventoryEntities.Product{
		ID:          product.ID.String(),
		Name:        product.Name,
		Description: product.Description,
		ImageURL:    product.ImageURL,
		Attributes:  (*json.RawMessage)(product.Attributes),
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}, nil
}
