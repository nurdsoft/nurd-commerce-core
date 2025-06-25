package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/repository"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	customerEntities "github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	sfEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
	shippingClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/client"
)

var (
	testCity        = "New York"
	testApartment   = "Apt 1"
	testPhoneNumber = "1234567890"
)

func Test_service_AddAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (
		*service, context.Context,
		*repository.MockRepository,
		*shippingClient.MockClient,
		*salesforce.MockClient,
	) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockShippingClient := shippingClient.NewMockClient(ctrl)
		mockSfClient := salesforce.NewMockClient(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:             mockRepo,
			log:              zap.NewExample().Sugar(),
			shippingClient:   mockShippingClient,
			salesforceClient: mockSfClient,
			customerClient:   customerclient.NewMockClient(ctrl),
		}
		return svc, ctx, mockRepo, mockShippingClient, mockSfClient
	}

	// TODO: Fix and enable
	// t.Run("Valid request with an address", func(t *testing.T) {
	// 	svc, ctx, mockRepo, mockShippingClient, mockSfClient := setup()
	// 	req := &entities.AddAddressRequest{
	// 		Address: &entities.AddressRequestBody{
	// 			FullName:    "John Doe",
	// 			Address:     "123 Main St",
	// 			City:        &testCity,
	// 			StateCode:   "NY",
	// 			PostalCode:  "10001",
	// 			Apartment:   &testApartment,
	// 			CountryCode: "US",
	// 			PhoneNumber: &testPhoneNumber,
	// 		},
	// 	}

	// 	expectedAddress := &entities.Address{
	// 		CustomerID:  uuid.MustParse(meta.XCustomerID(ctx)),
	// 		FullName:    req.Address.FullName,
	// 		Address:     req.Address.Address,
	// 		City:        req.Address.City,
	// 		StateCode:   req.Address.StateCode,
	// 		PostalCode:  req.Address.PostalCode,
	// 		Apartment:   req.Address.Apartment,
	// 		CountryCode: req.Address.CountryCode,
	// 		PhoneNumber: req.Address.PhoneNumber,
	// 	}

	// 	mockRepo.EXPECT().
	// 		CreateAddress(ctx, gomock.Any()).Return(expectedAddress, nil).Times(1)
	// 	mockShippingClient.EXPECT().
	// 		GetRatesEstimate(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

	// 	// Should this be here since this test function is testing AddAddress?
	// 	// mockRepo.EXPECT().
	// 	// 	FindByUUID(gomock.Any(), meta.XCustomerID(ctx)).Return(&entities.User{}, nil).AnyTimes()
	// 	svc.customerClient.(*customerclient.MockClient).EXPECT().
	// 		GetCustomerByID(gomock.Any(), meta.XCustomerID(ctx)).Return(&customerEntities.Customer{}, nil).AnyTimes()

	// 	mockSfClient.EXPECT().CreateUserAddress(gomock.Any(), gomock.Any()).Return(&sfEntities.CreateSFAddressResponse{}, nil).AnyTimes()
	// 	_, err := svc.AddAddress(ctx, req)

	// 	assert.NoError(t, err)
	// })

	t.Run("Valid request with an invalid address", func(t *testing.T) {
		svc, ctx, _, mockShippingClient, _ := setup()
		req := &entities.AddAddressRequest{
			Address: &entities.AddressRequestBody{
				FullName:    "John Doe",
				Address:     "123 Main St",
				City:        &testCity,
				StateCode:   "NY",
				PostalCode:  "TEST",
				Apartment:   &testApartment,
				CountryCode: "US",
				PhoneNumber: &testPhoneNumber,
			},
		}

		mockShippingClient.EXPECT().ValidateAddress(ctx, gomock.Any()).Return(nil, &appErrors.APIError{Message: "Invalid address"}).Times(1)
		_, err := svc.AddAddress(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid address")
	})

	t.Run("no user ID", func(t *testing.T) {
		svc, _, _, _, _ := setup()

		req := &entities.AddAddressRequest{
			Address: &entities.AddressRequestBody{
				FullName:    "John Doe",
				Address:     "123 Main St",
				City:        &testCity,
				StateCode:   "NY",
				PostalCode:  "10001",
				Apartment:   &testApartment,
				CountryCode: "US",
				PhoneNumber: &testPhoneNumber,
			},
		}
		ctx := meta.WithXCustomerID(context.Background(), "")
		_, err := svc.AddAddress(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_GetAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo := setup()

		mockRepo.EXPECT().GetAddresses(ctx, meta.XCustomerID(ctx)).Return([]entities.Address{}, nil).Times(1)
		_, err := svc.GetAddresses(ctx)

		assert.NoError(t, err)
	})

	t.Run("no user ID", func(t *testing.T) {
		svc, _, _ := setup()
		ctx := meta.WithXCustomerID(context.Background(), "")
		_, err := svc.GetAddresses(ctx)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_GetAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		addressID := uuid.New()
		req := &entities.GetAddressRequest{
			AddressID: addressID,
		}

		mockRepo.EXPECT().GetAddress(ctx, meta.XCustomerID(ctx), addressID.String()).Return(&entities.Address{}, nil).Times(1)
		_, err := svc.GetAddress(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("no user ID", func(t *testing.T) {
		svc, _, _ := setup()
		addressID := uuid.New()
		req := &entities.GetAddressRequest{
			AddressID: addressID,
		}

		ctx := meta.WithXCustomerID(context.Background(), "")
		_, err := svc.GetAddress(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_UpdateAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (
		*service, context.Context,
		*repository.MockRepository,
		*shippingClient.MockClient,
		*salesforce.MockClient,
	) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockShippingClient := shippingClient.NewMockClient(ctrl)
		mockSfClient := salesforce.NewMockClient(ctrl)

		userUUID := uuid.New()

		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:             mockRepo,
			log:              zap.NewExample().Sugar(),
			shippingClient:   mockShippingClient,
			salesforceClient: mockSfClient,
			customerClient:   customerclient.NewMockClient(ctrl),
		}
		return svc, ctx, mockRepo, mockShippingClient, mockSfClient
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo, mockShipingClient, mockSfClient := setup()
		addressID := uuid.New()
		req := &entities.UpdateAddressRequest{
			AddressID: addressID,
			Address: &entities.AddressRequestBody{
				FullName:    "John Doe",
				Address:     "123 Main St",
				City:        &testCity,
				StateCode:   "NY",
				PostalCode:  "10001",
				Apartment:   &testApartment,
				CountryCode: "US",
				PhoneNumber: &testPhoneNumber,
			},
		}

		expectedAddress := &entities.Address{
			ID:          addressID,
			CustomerID:  uuid.MustParse(meta.XCustomerID(ctx)),
			FullName:    req.Address.FullName,
			Address:     req.Address.Address,
			City:        req.Address.City,
			StateCode:   req.Address.StateCode,
			PostalCode:  req.Address.PostalCode,
			Apartment:   req.Address.Apartment,
			CountryCode: req.Address.CountryCode,
			PhoneNumber: req.Address.PhoneNumber,
		}

		mockRepo.EXPECT().UpdateAddress(ctx, expectedAddress).Return(&entities.Address{}, nil).Times(1)
		mockShipingClient.EXPECT().ValidateAddress(ctx, gomock.Any()).Return(nil, nil).Times(1)

		// sfID := "demo-sf-user-id"
		// mockRepo.EXPECT().FindByUUID(gomock.Any(), meta.XCustomerID(ctx)).Return(&entities.User{
		// 	SFUserID: &sfID,
		// }, nil).AnyTimes()
		customerID, _ := uuid.FromBytes([]byte(meta.XCustomerID(ctx)))
		svc.customerClient.(*customerclient.MockClient).EXPECT().
			GetCustomerByID(gomock.Any(), meta.XCustomerID(ctx)).Return(&customerEntities.Customer{
			ID: customerID,
		}, nil).AnyTimes()

		mockSfClient.EXPECT().UpdateUserAddress(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockSfClient.EXPECT().CreateUserAddress(gomock.Any(), gomock.Any()).Return(&sfEntities.CreateSFAddressResponse{}, nil).AnyTimes()
		mockRepo.EXPECT().UpdateAddressField(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		_, err := svc.UpdateAddress(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("no user ID", func(t *testing.T) {
		svc, _, _, _, _ := setup()
		addressID := uuid.New()
		req := &entities.UpdateAddressRequest{
			AddressID: addressID,
			Address: &entities.AddressRequestBody{
				FullName:    "John Doe",
				Address:     "123 Main St",
				City:        &testCity,
				StateCode:   "NY",
				PostalCode:  "10001",
				Apartment:   &testApartment,
				CountryCode: "US",
				PhoneNumber: &testPhoneNumber,
			},
		}
		ctx := meta.WithXCustomerID(context.Background(), "")
		_, err := svc.UpdateAddress(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("Valid request with an invalid address", func(t *testing.T) {
		svc, ctx, _, mockShippingClient, _ := setup()
		req := &entities.AddAddressRequest{
			Address: &entities.AddressRequestBody{
				FullName:    "John Doe",
				Address:     "123 Main St",
				City:        &testCity,
				StateCode:   "NY",
				PostalCode:  "TEST",
				Apartment:   &testApartment,
				CountryCode: "US",
				PhoneNumber: &testPhoneNumber,
			},
		}

		mockShippingClient.EXPECT().ValidateAddress(ctx, gomock.Any()).Return(nil, &appErrors.APIError{Message: "Invalid address"}).Times(1)
		_, err := svc.AddAddress(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid address")
	})
}

func Test_service_DeleteAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (
		*service, context.Context,
		*repository.MockRepository,
		*salesforce.MockClient,
	) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockSfClient := salesforce.NewMockClient(ctrl)

		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:             mockRepo,
			log:              zap.NewExample().Sugar(),
			salesforceClient: mockSfClient,
		}
		return svc, ctx, mockRepo, mockSfClient
	}

	t.Run("Valid request", func(t *testing.T) {
		svc, ctx, mockRepo, mockSfClient := setup()
		addressID := uuid.New()
		req := &entities.DeleteAddressRequest{
			AddressID: addressID,
		}
		mockRepo.EXPECT().GetAddress(ctx, meta.XCustomerID(ctx), addressID.String()).Return(&entities.Address{
			ID: addressID,
		}, nil).Times(1)
		mockRepo.EXPECT().DeleteAddress(ctx, meta.XCustomerID(ctx), addressID.String()).Return(nil).Times(1)
		mockSfClient.EXPECT().DeleteUserAddress(gomock.Any(), addressID).Return(nil).AnyTimes()
		err := svc.DeleteAddress(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("no user ID", func(t *testing.T) {
		svc, _, _, _ := setup()
		addressID := uuid.New()
		req := &entities.DeleteAddressRequest{
			AddressID: addressID,
		}
		ctx := meta.WithXCustomerID(context.Background(), "")
		err := svc.DeleteAddress(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

// Note: I believe this test should be in the stripe client package
// func Test_service_getUserStripeId(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	setup := func() (*service, context.Context, *repository.MockRepository, *stripeClient.MockClient) {
// 		mockRepo := repository.NewMockRepository(ctrl)
// 		mockStripeClient := stripeClient.NewMockClient(ctrl)
// 		userUUID := uuid.New()
// 		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
// 		svc := &service{
// 			repo:         mockRepo,
// 			stripeClient: mockStripeClient,
// 			log:          zap.NewExample().Sugar(),
// 		}
// 		return svc, ctx, mockRepo, mockStripeClient
// 	}

// 	t.Run("user not found in repository", func(t *testing.T) {
// 		svc, ctx, mockRepo, _ := setup()
// 		mockRepo.EXPECT().FindByUUID(ctx, meta.XCustomerID(ctx)).Return(nil, &appErrors.APIError{Message: "user not found"})

// 		_, _, err := svc.getUserStripeId(ctx, meta.XCustomerID(ctx))
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "user not found")
// 	})

// 	t.Run("creates new Stripe customer", func(t *testing.T) {
// 		svc, ctx, mockRepo, mockStripeClient := setup()

// 		user := &entities.User{
// 			UserUUID:  uuid.New(),
// 			FirstName: "John",
// 			Email:     "john.doe@example.com",
// 		}
// 		mockRepo.EXPECT().FindByUUID(ctx, meta.XCustomerID(ctx)).Return(user, nil)
// 		mockStripeClient.EXPECT().CreateCustomer(ctx, gomock.Any()).Return(&stripeEntities.CreateCustomerResponse{Id: "cust_123"}, nil)
// 		mockRepo.EXPECT().Update(ctx, gomock.Any(), meta.XCustomerID(ctx)).Return(nil)

// 		stripeId, created, err := svc.getUserStripeId(ctx, meta.XCustomerID(ctx))
// 		assert.NoError(t, err)
// 		assert.True(t, created)
// 		assert.Equal(t, "cust_123", *stripeId)
// 	})

// 	t.Run("fails to update repository after creating Stripe customer", func(t *testing.T) {
// 		svc, ctx, mockRepo, mockStripeClient := setup()

// 		user := &entities.User{
// 			UserUUID:  uuid.New(),
// 			FirstName: "John",
// 			Email:     "john.doe@example.com",
// 		}
// 		mockRepo.EXPECT().FindByUUID(ctx, meta.XCustomerID(ctx)).Return(user, nil)
// 		mockStripeClient.EXPECT().CreateCustomer(ctx, gomock.Any()).Return(&stripeEntities.CreateCustomerResponse{Id: "cust_123"}, nil)
// 		mockRepo.EXPECT().Update(ctx, gomock.Any(), meta.XCustomerID(ctx)).Return(&appErrors.APIError{Message: "update failed"})

// 		_, _, err := svc.getUserStripeId(ctx, meta.XCustomerID(ctx))
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "update failed")
// 	})

// 	t.Run("user already has Stripe ID", func(t *testing.T) {
// 		svc, ctx, mockRepo, _ := setup()

// 		stripeId := "cust_123"
// 		user := &entities.User{
// 			UserUUID: uuid.New(),
// 			StripeId: &stripeId,
// 		}
// 		mockRepo.EXPECT().FindByUUID(ctx, meta.XCustomerID(ctx)).Return(user, nil)

// 		resultStripeId, created, err := svc.getUserStripeId(ctx, meta.XCustomerID(ctx))
// 		assert.NoError(t, err)
// 		assert.False(t, created)
// 		assert.Equal(t, "cust_123", *resultStripeId)
// 	})
// }
