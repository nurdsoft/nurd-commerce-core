package client

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/service"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetAccountByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

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
		expectedAccount := &entities.Account{ID: accountID, Name: "Test Account"}
		mockService.EXPECT().GetAccountByID(ctx, accountID).Return(expectedAccount, nil)

		account, err := client.GetAccountByID(ctx, accountID)
		assert.NoError(t, err)
		assert.Equal(t, expectedAccount, account)
	})
}

func TestClient_CreateUserAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.CreateSFUserRequest{FirstName: "John", LastName: "Doe", PersonEmail: "john.doe@example.com"}

	t.Run("Error creating user account", func(t *testing.T) {
		mockService.EXPECT().CreateUserAccount(ctx, req).Return(nil, &appErrors.APIError{Message: "Error creating user account"})

		_, err := client.CreateUserAccount(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating user account", err.Error())
	})

	t.Run("Valid user account creation", func(t *testing.T) {
		expectedResponse := &entities.CreateSFUserResponse{ID: "0011N00001Gv7PQQAZ"}
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
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.UpdateSFUserRequest{ID: "0011N00001Gv7PQQAZ", FirstName: "John", LastName: "Doe"}

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
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.CreateSFAddressRequest{AccountC: "0011N00001Gv7PQQAZ", ShippingStreetC: "123 Main St"}

	t.Run("Error creating user address", func(t *testing.T) {
		mockService.EXPECT().CreateUserAddress(ctx, req).Return(nil, &appErrors.APIError{Message: "Error creating user address"})

		_, err := client.CreateUserAddress(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating user address", err.Error())
	})

	t.Run("Valid user address creation", func(t *testing.T) {
		expectedResponse := &entities.CreateSFAddressResponse{ID: "0011N00001Gv7PQQAZ"}
		mockService.EXPECT().CreateUserAddress(ctx, req).Return(expectedResponse, nil)

		response, err := client.CreateUserAddress(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_UpdateUserAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.UpdateSFAddressRequest{AccountC: "0011N00001Gv7PQQAZ", ShippingStreetC: "123 Main St"}

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
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

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
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.CreateSFProductRequest{Name: "Test Product"}

	t.Run("Error creating product", func(t *testing.T) {
		mockService.EXPECT().CreateProduct(ctx, req).Return(nil, &appErrors.APIError{Message: "Error creating product"})

		_, err := client.CreateProduct(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating product", err.Error())
	})

	t.Run("Valid product creation", func(t *testing.T) {
		expectedResponse := &entities.CreateSFProductResponse{ID: "0011N00001Gv7PQQAZ"}
		mockService.EXPECT().CreateProduct(ctx, req).Return(expectedResponse, nil)

		response, err := client.CreateProduct(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_CreatePriceBookEntry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.CreateSFPriceBookEntryRequest{Product2ID: "0011N00001Gv7PQQAZ", UnitPrice: 100}

	t.Run("Error creating price book entry", func(t *testing.T) {
		mockService.EXPECT().CreatePriceBookEntry(ctx, req).Return(nil, &appErrors.APIError{Message: "Error creating price book entry"})

		_, err := client.CreatePriceBookEntry(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating price book entry", err.Error())
	})

	t.Run("Valid price book entry creation", func(t *testing.T) {
		expectedResponse := &entities.CreateSFPriceBookEntryResponse{ID: "0011N00001Gv7PQQAZ"}
		mockService.EXPECT().CreatePriceBookEntry(ctx, req).Return(expectedResponse, nil)

		response, err := client.CreatePriceBookEntry(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_CreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.CreateSFOrderRequest{AccountID: "0011N00001Gv7PQQAZ", OrderReferenceC: "ORD123"}

	t.Run("Error creating order", func(t *testing.T) {
		mockService.EXPECT().CreateOrder(ctx, req).Return(nil, &appErrors.APIError{Message: "Error creating order"})

		_, err := client.CreateOrder(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error creating order", err.Error())
	})

	t.Run("Valid order creation", func(t *testing.T) {
		expectedResponse := &entities.CreateSFOrderResponse{ID: "0011N00001Gv7PQQAZ"}
		mockService.EXPECT().CreateOrder(ctx, req).Return(expectedResponse, nil)

		response, err := client.CreateOrder(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_AddOrderItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	items := []*entities.OrderItem{{OrderID: "0011N00001Gv7PQQAZ", Quantity: 1, UnitPrice: 100}}

	t.Run("Error adding order items", func(t *testing.T) {
		mockService.EXPECT().AddOrderItems(ctx, items).Return(nil, &appErrors.APIError{Message: "Error adding order items"})

		_, err := client.AddOrderItems(ctx, items)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error adding order items", err.Error())
	})

	t.Run("Valid order items addition", func(t *testing.T) {
		expectedResponse := &entities.AddOrderItemResponse{HasErrors: false}
		mockService.EXPECT().AddOrderItems(ctx, items).Return(expectedResponse, nil)

		response, err := client.AddOrderItems(ctx, items)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}

func TestClient_UpdateOrderStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := &entities.UpdateOrderRequest{OrderId: "0011N00001Gv7PQQAZ", Status: "Completed"}

	t.Run("Error updating order status", func(t *testing.T) {
		mockService.EXPECT().UpdateOrderStatus(ctx, req).Return(&appErrors.APIError{Message: "Error updating order status"})

		err := client.UpdateOrderStatus(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Equal(t, "Error updating order status", err.Error())
	})

	t.Run("Valid order status update", func(t *testing.T) {
		mockService.EXPECT().UpdateOrderStatus(ctx, req).Return(nil)

		err := client.UpdateOrderStatus(ctx, req)
		assert.NoError(t, err)
	})
}

func TestClient_GetOrderItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

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
		expectedResponse := &entities.GetOrderItemsResponse{Records: []entities.Records{{ID: "0011N00001Gv7PQQAZ"}}}
		mockService.EXPECT().GetOrderItems(ctx, orderID).Return(expectedResponse, nil)

		response, err := client.GetOrderItems(ctx, orderID)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})
}
