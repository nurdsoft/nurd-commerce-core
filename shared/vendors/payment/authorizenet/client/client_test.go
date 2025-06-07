package client

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestClient_CreateCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := entities.CreateCustomerRequest{
		CustomerID:  "cust_123",
		Description: "Test Customer",
		Email:       "test@example.com",
	}

	t.Run("Success", func(t *testing.T) {
		expectedResp := entities.CreateCustomerResponse{
			ProfileID: "profile_123",
		}
		mockService.EXPECT().
			CreateCustomerProfile(ctx, req).Return(expectedResp, nil)

		resp, err := client.CreateCustomer(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("Error", func(t *testing.T) {
		mockService.EXPECT().
			CreateCustomerProfile(ctx, req).Return(entities.CreateCustomerResponse{}, errors.New("service error"))

		resp, err := client.CreateCustomer(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, entities.CreateCustomerResponse{}, resp)
		assert.Equal(t, "service error", err.Error())
	})
}

func TestClient_CreateCustomerPaymentProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := entities.CreateCustomerPaymentProfileRequest{
		ProfileID:      "profile_123",
		CardNumber:     "4111111111111111",
		ExpirationDate: "2025-12",
	}

	t.Run("Success", func(t *testing.T) {
		expectedResp := entities.CreateCustomerPaymentProfileResponse{
			ProfileID:        "profile_123",
			PaymentProfileID: "payprof_123",
		}
		mockService.EXPECT().
			CreateCustomerPaymentProfile(ctx, req).Return(expectedResp, nil)

		resp, err := client.CreateCustomerPaymentProfile(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("Error", func(t *testing.T) {
		mockService.EXPECT().
			CreateCustomerPaymentProfile(ctx, req).Return(entities.CreateCustomerPaymentProfileResponse{}, errors.New("service error"))

		resp, err := client.CreateCustomerPaymentProfile(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, entities.CreateCustomerPaymentProfileResponse{}, resp)
		assert.Equal(t, "service error", err.Error())
	})
}

func TestClient_GetCustomerPaymentMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := entities.GetPaymentProfilesRequest{
		ProfileID: "profile_123",
	}

	t.Run("Success", func(t *testing.T) {
		expectedResp := entities.GetPaymentProfilesResponse{
			PaymentProfiles: []entities.PaymentProfile{{
				ID:             "payprof_123",
				CardNumber:     "411111******1111",
				CardType:       "Visa",
				ExpirationDate: "2025-12",
			}},
		}
		mockService.EXPECT().
			GetCustomerPaymentProfiles(ctx, req).Return(expectedResp, nil)

		resp, err := client.GetCustomerPaymentMethods(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("Error", func(t *testing.T) {
		mockService.EXPECT().
			GetCustomerPaymentProfiles(ctx, req).Return(entities.GetPaymentProfilesResponse{}, errors.New("service error"))

		resp, err := client.GetCustomerPaymentMethods(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, entities.GetPaymentProfilesResponse{}, resp)
		assert.Equal(t, "service error", err.Error())
	})
}

func TestClient_CreatePayment(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)

	ctx := context.Background()
	req := entities.CreatePaymentTransactionRequest{
		Amount:    decimal.NewFromInt(1000),
		ProfileID: "profile_123",
	}

	t.Run("Success: Approved", func(t *testing.T) {
		svcResp := entities.CreatePaymentTransactionResponse{
			ID:     "txn_123",
			Status: "approved",
		}
		mockService.EXPECT().CreatePaymentTransaction(ctx, req).Return(svcResp, nil)

		resp, err := client.CreatePayment(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, providers.PaymentProviderResponse{
			ID:     svcResp.ID,
			Status: providers.PaymentStatusSuccess,
		}, resp)
	})

	t.Run("Success: Declined", func(t *testing.T) {
		svcResp := entities.CreatePaymentTransactionResponse{
			ID:     "txn_123",
			Status: "declined",
		}
		mockService.EXPECT().CreatePaymentTransaction(ctx, req).Return(svcResp, nil)

		resp, err := client.CreatePayment(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, providers.PaymentProviderResponse{
			ID:     svcResp.ID,
			Status: providers.PaymentStatusFailed,
		}, resp)
	})

	t.Run("Error", func(t *testing.T) {
		mockService.EXPECT().CreatePaymentTransaction(ctx, req).Return(entities.CreatePaymentTransactionResponse{}, errors.New("service error"))

		resp, err := client.CreatePayment(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, providers.PaymentProviderResponse{}, resp)
		assert.Equal(t, "service error", err.Error())
	})

	t.Run("Invalid request type", func(t *testing.T) {
		resp, err := client.CreatePayment(ctx, "invalid type")

		assert.Error(t, err)
		assert.Equal(t, providers.PaymentProviderResponse{}, resp)
		assert.Equal(t, "invalid payment request type", err.Error())
	})
}
