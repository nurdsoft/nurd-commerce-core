package client

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	customerEntities "github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	orderEntities "github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	ordersrepo "github.com/nurdsoft/nurd-commerce-core/internal/orders/repository"
	productEntities "github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/providers"
	sfEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetAccountByID(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	accountID := "0011N00001Gv7PQQAZ"

	t.Run("Account not found", func(t *testing.T) {
		mockService.EXPECT().GetAccountByID(ctx, accountID).Return(nil, &appErrors.APIError{Message: "Account not found"})

		_, err := client.GetAccountByID(ctx, accountID)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Account not found", err.Error())
	})

	t.Run("Valid account ID", func(t *testing.T) {
		expectedAccount := &sfEntities.Account{ID: accountID, Name: "Test Account"}
		mockService.EXPECT().GetAccountByID(ctx, accountID).Return(expectedAccount, nil)

		account, err := client.GetAccountByID(ctx, accountID)
		assert.NoError(t, err)
		assert.Equal(t, expectedAccount, account)
	})
}

func TestClient_CreateUserAccount(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	req := &sfEntities.CreateSFUserRequest{FirstName: "John", LastName: "Doe", PersonEmail: "john.doe@example.com"}

	t.Run("Error creating user account", func(t *testing.T) {
		mockService.EXPECT().CreateUserAccount(ctx, req).Return(nil, &appErrors.APIError{Message: "Error creating user account"})

		_, err := client.CreateUserAccount(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating user account", err.Error())
	})

	t.Run("Valid user account creation", func(t *testing.T) {
		expectedResponse := &sfEntities.CreateSFUserResponse{ID: "0011N00001Gv7PQQAZ"}
		mockService.EXPECT().CreateUserAccount(ctx, req).Return(expectedResponse, nil)

		response, err := client.CreateUserAccount(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_UpdateUserAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	req := &sfEntities.UpdateSFUserRequest{ID: "0011N00001Gv7PQQAZ", FirstName: "John", LastName: "Doe"}

	t.Run("Error updating user account", func(t *testing.T) {
		mockService.EXPECT().UpdateUserAccount(ctx, req).Return(&appErrors.APIError{Message: "Error updating user account"})

		err := client.UpdateUserAccount(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error updating user account", err.Error())
	})

	t.Run("Valid user account update", func(t *testing.T) {
		mockService.EXPECT().UpdateUserAccount(ctx, req).Return(nil)

		err := client.UpdateUserAccount(ctx, req)
		assert.NoError(t, err)
	})
}

func TestClient_CreateUserAddress(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	req := &sfEntities.CreateSFAddressRequest{AccountC: "0011N00001Gv7PQQAZ", ShippingStreetC: "123 Main St"}

	t.Run("Error creating user address", func(t *testing.T) {
		mockService.EXPECT().CreateUserAddress(ctx, req).Return(nil, &appErrors.APIError{Message: "Error creating user address"})

		_, err := client.CreateUserAddress(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating user address", err.Error())
	})

	t.Run("Valid user address creation", func(t *testing.T) {
		expectedResponse := &sfEntities.CreateSFAddressResponse{ID: "0011N00001Gv7PQQAZ"}
		mockService.EXPECT().CreateUserAddress(ctx, req).Return(expectedResponse, nil)

		response, err := client.CreateUserAddress(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_UpdateUserAddress(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	req := &sfEntities.UpdateSFAddressRequest{AccountC: "0011N00001Gv7PQQAZ", ShippingStreetC: "123 Main St"}

	t.Run("Error updating user address", func(t *testing.T) {
		mockService.EXPECT().UpdateUserAddress(ctx, req).Return(&appErrors.APIError{Message: "Error updating user address"})

		err := client.UpdateUserAddress(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error updating user address", err.Error())
	})

	t.Run("Valid user address update", func(t *testing.T) {
		mockService.EXPECT().UpdateUserAddress(ctx, req).Return(nil)

		err := client.UpdateUserAddress(ctx, req)
		assert.NoError(t, err)
	})
}

func TestClient_DeleteUserAddress(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	addressID := "0011N00001Gv7PQQAZ"

	t.Run("Error deleting user address", func(t *testing.T) {
		mockService.EXPECT().DeleteUserAddress(ctx, addressID).Return(&appErrors.APIError{Message: "Error deleting user address"})

		err := client.DeleteUserAddress(ctx, addressID)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error deleting user address", err.Error())
	})

	t.Run("Valid user address deletion", func(t *testing.T) {
		mockService.EXPECT().DeleteUserAddress(ctx, addressID).Return(nil)

		err := client.DeleteUserAddress(ctx, addressID)
		assert.NoError(t, err)
	})
}

func TestClient_CreateProduct(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	req := &sfEntities.CreateSFProductRequest{Name: "Test Product"}

	t.Run("Error creating product", func(t *testing.T) {
		mockService.EXPECT().CreateProduct(ctx, req).Return(nil, &appErrors.APIError{Message: "Error creating product"})

		_, err := client.CreateProduct(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating product", err.Error())
	})

	t.Run("Valid product creation", func(t *testing.T) {
		expectedResponse := &sfEntities.CreateSFProductResponse{ID: "0011N00001Gv7PQQAZ"}
		mockService.EXPECT().CreateProduct(ctx, req).Return(expectedResponse, nil)

		response, err := client.CreateProduct(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_CreatePriceBookEntry(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	req := &sfEntities.CreateSFPriceBookEntryRequest{Product2ID: "0011N00001Gv7PQQAZ", UnitPrice: 100}

	t.Run("Error creating price book entry", func(t *testing.T) {
		mockService.EXPECT().CreatePriceBookEntry(ctx, req).Return(nil, &appErrors.APIError{Message: "Error creating price book entry"})

		_, err := client.CreatePriceBookEntry(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating price book entry", err.Error())
	})

	t.Run("Valid price book entry creation", func(t *testing.T) {
		expectedResponse := &sfEntities.CreateSFPriceBookEntryResponse{ID: "0011N00001Gv7PQQAZ"}
		mockService.EXPECT().CreatePriceBookEntry(ctx, req).Return(expectedResponse, nil)

		response, err := client.CreatePriceBookEntry(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_CreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	orderID := uuid.New()
	customerID := uuid.New()
	productID := uuid.New()
	customerSalesforceID := uuid.NewString()
	description := "Test Item"
	testCarrierName := "Test Carrier"
	testCarrierService := "Test Service"
	testShippingRate := decimal.NewFromInt(10)
	testEstimatedDeliveryDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	req := entities.CreateInventoryOrderRequest{
		Order: orderEntities.Order{
			ID:                            orderID,
			OrderReference:                "ORD123",
			Status:                        orderEntities.Pending,
			Total:                         decimal.NewFromInt(100),
			Subtotal:                      decimal.NewFromInt(100),
			ShippingRate:                  &testShippingRate,
			TaxAmount:                     decimal.NewFromInt(10),
			ShippingCarrierName:           &testCarrierName,
			ShippingServiceType:           &testCarrierService,
			Currency:                      "USD",
			CustomerID:                    customerID,
			CartID:                        uuid.New(),
			ShippingEstimatedDeliveryDate: &testEstimatedDeliveryDate,
		},
		OrderItems: []*orderEntities.OrderItem{
			{
				ID:          uuid.New(),
				OrderID:     orderID,
				Quantity:    1,
				ProductID:   productID,
				SKU:         "SKU123",
				Name:        "Test Item",
				Description: &description,
			},
		},
		CartItems: []cartEntities.CartItemDetail{
			{
				ID:        uuid.New(),
				ProductID: productID,
				Quantity:  1,
				Price:     decimal.NewFromInt(100),
				Name:      "Test Item",
			},
		},
		Customer: customerEntities.Customer{
			ID:           customerID,
			SalesforceID: &customerSalesforceID,
		},
	}

	t.Run("Error creating order", func(t *testing.T) {
		mockService.EXPECT().CreateOrder(ctx, gomock.Any()).Return(nil, &appErrors.APIError{Message: "Error creating order"})

		_, err := client.CreateOrder(ctx, req)

		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating order", err.Error())
	})

	t.Run("Valid order creation", func(t *testing.T) {
		salesforceOrderID := "0011N00001Gv7PQQAZ"
		salesforcePricebookEntryID := "0011N00001Gv7PQQAZ"
		mockService.EXPECT().
			CreateOrder(ctx, gomock.Any()).
			Do(func(_ context.Context, sfReq *sfEntities.CreateSFOrderRequest) {
				assert.Equal(t, req.Order.OrderReference, sfReq.OrderReferenceC)
				assert.Equal(t, *req.Customer.SalesforceID, sfReq.AccountID)
				assert.Equal(t, req.Order.Status.String(), sfReq.Status)
				assert.Equal(t, req.Order.Total.String(), sfReq.TotalC)
				assert.Equal(t, req.Order.Subtotal.String(), sfReq.SubTotalC)
				assert.Equal(t, req.Order.ShippingRate.String(), sfReq.ShippingRateC)
				assert.Equal(t, req.Order.TaxAmount.String(), sfReq.TaxAmountC)
				assert.Equal(t, req.Order.Currency, sfReq.CurrencyC)
				assert.Equal(t, req.Order.ShippingCarrierName, &sfReq.ShippingCarrierNameC)
				assert.Equal(t, req.Order.ShippingServiceType, &sfReq.ShippingCarrierServiceC)
				assert.Equal(t, req.Order.ShippingEstimatedDeliveryDate.Format("2006-01-02"), sfReq.EstimatedDeliveryDateC)
			}).
			Return(&sfEntities.CreateSFOrderResponse{ID: salesforceOrderID, Success: true}, nil)

		mockOrdersRepo.EXPECT().
			Update(ctx, gomock.Any(), orderID.String(), customerID.String()).
			Do(func(_ context.Context, data map[string]interface{}, orderID string, customerID string) {
				assert.Equal(t, salesforceOrderID, data["salesforce_id"])
			}).
			Return(nil)

		mockProductClient.EXPECT().
			GetProductsByIDs(ctx, []string{productID.String()}).Return([]productEntities.Product{
			{
				ID:                         productID,
				Name:                       "Test Item",
				SalesforcePricebookEntryId: &salesforcePricebookEntryID,
			},
		}, nil)

		mockService.EXPECT().
			AddOrderItems(ctx, gomock.Any()).
			Do(func(_ context.Context, items []*sfEntities.OrderItem) {
				assert.Equal(t, salesforceOrderID, items[0].OrderID)
				assert.Equal(t, 1, items[0].Quantity)
				assert.Equal(t, float64(100), items[0].UnitPrice)
			}).
			Return(&sfEntities.AddOrderItemResponse{HasErrors: false}, nil)
		salesforceItems := []sfEntities.Records{
			{
				ID:          salesforceOrderID,
				TypeC:       *req.OrderItems[0].Description,
				Description: *req.OrderItems[0].Description,
				Product2ID:  productID.String(),
			},
		}
		mockService.EXPECT().
			GetOrderItems(ctx, salesforceOrderID).
			Return(&sfEntities.GetOrderItemsResponse{Records: salesforceItems}, nil)
		mockOrdersRepo.EXPECT().
			AddSalesforceIDPerOrderItem(ctx, gomock.Any()).
			Do(func(_ context.Context, data map[string]string) {
				assert.Len(t, data, 1)
				assert.Equal(t, salesforceItems[0].ID, data[req.OrderItems[0].ID.String()])
			}).
			Return(nil)

		_, err := client.CreateOrder(ctx, req)
		assert.NoError(t, err)
	})
}

func TestClient_AddOrderItems(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	items := []*sfEntities.OrderItem{{OrderID: "0011N00001Gv7PQQAZ", Quantity: 1, UnitPrice: 100}}

	t.Run("Error adding order items", func(t *testing.T) {
		mockService.EXPECT().AddOrderItems(ctx, items).Return(nil, &appErrors.APIError{Message: "Error adding order items"})

		_, err := client.AddOrderItems(ctx, items)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error adding order items", err.Error())
	})

	t.Run("Valid order items addition", func(t *testing.T) {
		expectedResponse := &sfEntities.AddOrderItemResponse{HasErrors: false}
		mockService.EXPECT().AddOrderItems(ctx, items).Return(expectedResponse, nil)

		response, err := client.AddOrderItems(ctx, items)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_UpdateOrderStatus(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	customerSalesforceID := "0011N00001Gv7PQQAZ"
	req := entities.UpdateInventoryOrderStatusRequest{
		Order:    orderEntities.Order{SalesforceID: "0011N00001Gv7PQQAZ"},
		Customer: customerEntities.Customer{SalesforceID: &customerSalesforceID},
		Status:   "Completed",
	}
	expectedReq := &sfEntities.UpdateOrderRequest{OrderId: "0011N00001Gv7PQQAZ", AccountID: customerSalesforceID, Status: "Completed"}

	t.Run("Error updating order status", func(t *testing.T) {
		mockService.EXPECT().UpdateOrderStatus(ctx, expectedReq).Return(&appErrors.APIError{Message: "Error updating order status"})

		err := client.UpdateOrderStatus(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error updating order status", err.Error())
	})

	t.Run("Valid order status update", func(t *testing.T) {
		mockService.EXPECT().UpdateOrderStatus(ctx, expectedReq).Return(nil)

		err := client.UpdateOrderStatus(ctx, req)
		assert.NoError(t, err)
	})
}

func TestClient_GetOrderItems(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockService(ctrl)
	mockProductClient := productclient.NewMockClient(ctrl)
	mockOrdersRepo := ordersrepo.NewMockRepository(ctrl)
	client := NewClient(mockService, providers.ProviderSalesforce, mockProductClient, mockOrdersRepo)

	ctx := context.Background()
	orderID := "0011N00001Gv7PQQAZ"

	t.Run("Error getting order items", func(t *testing.T) {
		mockService.EXPECT().GetOrderItems(ctx, orderID).Return(nil, &appErrors.APIError{Message: "Error getting order items"})

		_, err := client.GetOrderItems(ctx, orderID)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error getting order items", err.Error())
	})

	t.Run("Valid order items retrieval", func(t *testing.T) {
		expectedResponse := &sfEntities.GetOrderItemsResponse{Records: []sfEntities.Records{{ID: "0011N00001Gv7PQQAZ"}}}
		mockService.EXPECT().GetOrderItems(ctx, orderID).Return(expectedResponse, nil)

		response, err := client.GetOrderItems(ctx, orderID)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}
