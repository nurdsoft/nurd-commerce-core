package service

import (
	"context"
	"errors"
	"testing"

	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/nullable"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/repository"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/providers"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	salesforceEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockRepo := repository.NewMockRepository(ctrl)
	mockLogger := zap.NewExample().Sugar()
	mockConfig := cfg.Config{}
	mockSfClient := salesforce.NewMockClient(ctrl)
	mockInventoryClient := inventory.NewMockClient(ctrl)

	// Call the constructor
	svc := New(mockRepo, mockLogger, mockConfig, mockSfClient, mockInventoryClient)

	// Verify the service was created and is not nil
	assert.NotNil(t, svc, "Service should not be nil")

	// Type assertion to verify it's the correct type
	_, ok := svc.(*service)
	assert.True(t, ok, "Service should be of type *service")
}

func Test_service_CreateCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *inventory.MockClient, *salesforce.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockInventoryClient := inventory.NewMockClient(ctrl)
		mockSfClient := salesforce.NewMockClient(ctrl)
		ctx := context.Background()
		svc := &service{
			repo:            mockRepo,
			log:             zap.NewExample().Sugar(),
			config:          cfg.Config{},
			sfClient:        mockSfClient,
			inventoryClient: mockInventoryClient,
		}
		return svc, ctx, mockRepo, mockInventoryClient, mockSfClient
	}

	t.Run("Successfully creates customer without predefined ID", func(t *testing.T) {
		svc, ctx, mockRepo, mockInventoryClient, mockSfClient := setup()

		firstName := "John"
		lastName := "Doe"
		email := "john.doe@example.com"
		phoneNumber := "1234567890"

		req := &entities.CreateCustomerRequest{
			Data: &entities.CreateCustomerRequestBody{
				FirstName:   firstName,
				LastName:    &lastName,
				Email:       email,
				PhoneNumber: &phoneNumber,
			},
		}

		expectedCustomer := &entities.Customer{
			FirstName:   firstName,
			LastName:    &lastName,
			Email:       email,
			PhoneNumber: &phoneNumber,
		}

		mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, customer *entities.Customer) (*entities.Customer, error) {
			// Verify that a UUID was generated
			assert.NotEqual(t, uuid.Nil, customer.ID)
			expectedCustomer.ID = customer.ID
			return expectedCustomer, nil
		}).Times(1)

		mockInventoryClient.EXPECT().GetProvider().Return(providers.ProviderSalesforce).Times(1)

		// Mock the Salesforce call that happens in the goroutine
		mockSfClient.EXPECT().CreateUserAccount(gomock.Any(), gomock.Any()).Return(&salesforceEntities.CreateSFUserResponse{
			ID: "sf_user_123",
		}, nil).Times(1)
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		result, err := svc.CreateCustomer(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer, result)
	})

	t.Run("Successfully creates customer with predefined ID", func(t *testing.T) {
		svc, ctx, mockRepo, mockInventoryClient, mockSfClient := setup()

		customerID := uuid.New()
		firstName := "John"
		email := "john.doe@example.com"

		req := &entities.CreateCustomerRequest{
			Data: &entities.CreateCustomerRequestBody{
				ID:        &customerID,
				FirstName: firstName,
				Email:     email,
			},
		}

		expectedCustomer := &entities.Customer{
			ID:        customerID,
			FirstName: firstName,
			Email:     email,
		}

		mockRepo.EXPECT().Create(ctx, expectedCustomer).Return(expectedCustomer, nil).Times(1)
		mockInventoryClient.EXPECT().GetProvider().Return(providers.ProviderSalesforce).Times(1)

		// Mock the Salesforce call that happens in the goroutine
		mockSfClient.EXPECT().CreateUserAccount(gomock.Any(), gomock.Any()).Return(&salesforceEntities.CreateSFUserResponse{
			ID: "sf_user_123",
		}, nil).Times(1)
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		result, err := svc.CreateCustomer(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer, result)
	})

	t.Run("Repository error is propagated", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()

		req := &entities.CreateCustomerRequest{
			Data: &entities.CreateCustomerRequestBody{
				FirstName: "John",
				Email:     "john.doe@example.com",
			},
		}

		expectedErr := errors.New("database error")
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, expectedErr).Times(1)

		result, err := svc.CreateCustomer(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("Creates customer with non-Salesforce provider", func(t *testing.T) {
		svc, ctx, mockRepo, mockInventoryClient, _ := setup()

		req := &entities.CreateCustomerRequest{
			Data: &entities.CreateCustomerRequestBody{
				FirstName: "John",
				Email:     "john.doe@example.com",
			},
		}

		expectedCustomer := &entities.Customer{
			FirstName: "John",
			Email:     "john.doe@example.com",
		}

		mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, customer *entities.Customer) (*entities.Customer, error) {
			expectedCustomer.ID = customer.ID
			return expectedCustomer, nil
		}).Times(1)

		mockInventoryClient.EXPECT().GetProvider().Return(providers.ProviderType("other-provider")).Times(1)

		result, err := svc.CreateCustomer(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer, result)
	})

	// t.Run("Creates customer with empty last name uses zero-width space", func(t *testing.T) {
	// 	svc, ctx, mockRepo, mockInventoryClient, mockSfClient := setup()

	// 	firstName := "John"
	// 	lastName := ""
	// 	email := "john.doe@example.com"

	// 	req := &entities.CreateCustomerRequest{
	// 		Data: &entities.CreateCustomerRequestBody{
	// 			FirstName: firstName,
	// 			LastName:  &lastName,
	// 			Email:     email,
	// 		},
	// 	}

	// 	expectedCustomer := &entities.Customer{
	// 		FirstName: firstName,
	// 		LastName:  &lastName,
	// 		Email:     email,
	// 	}

	// 	mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, customer *entities.Customer) (*entities.Customer, error) {
	// 		expectedCustomer.ID = customer.ID
	// 		return expectedCustomer, nil
	// 	}).Times(1)

	// 	mockInventoryClient.EXPECT().GetProvider().Return(providers.ProviderSalesforce).Times(1)

	// 	// Mock the Salesforce call that happens in the goroutine
	// 	mockSfClient.EXPECT().CreateUserAccount(gomock.Any(), gomock.Any()).Return(&salesforceEntities.CreateSFUserResponse{
	// 		ID: "sf_user_123",
	// 	}, nil).Times(1)
	// 	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	// 	result, err := svc.CreateCustomer(ctx, req)
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, expectedCustomer, result)
	// })
}

func Test_service_GetCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		customerUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), customerUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("Successfully gets customer", func(t *testing.T) {
		svc, ctx, mockRepo := setup()

		expectedCustomer := &entities.Customer{
			ID:        uuid.MustParse(meta.XCustomerID(ctx)),
			FirstName: "John",
			Email:     "john.doe@example.com",
		}

		mockRepo.EXPECT().FindByID(ctx, meta.XCustomerID(ctx)).Return(expectedCustomer, nil).Times(1)

		result, err := svc.GetCustomer(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer, result)
	})

	t.Run("No customer ID in context", func(t *testing.T) {
		svc, _, _ := setup()
		ctx := meta.WithXCustomerID(context.Background(), "")

		_, err := svc.GetCustomer(ctx)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("Repository error is propagated", func(t *testing.T) {
		svc, ctx, mockRepo := setup()

		expectedErr := errors.New("database error")
		mockRepo.EXPECT().FindByID(ctx, meta.XCustomerID(ctx)).Return(nil, expectedErr).Times(1)

		result, err := svc.GetCustomer(ctx)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
}

func Test_service_GetCustomerByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, mockRepo
	}

	t.Run("Successfully gets customer by ID", func(t *testing.T) {
		svc, mockRepo := setup()
		ctx := context.Background()
		customerID := uuid.New().String()

		expectedCustomer := &entities.Customer{
			ID:        uuid.MustParse(customerID),
			FirstName: "John",
			Email:     "john.doe@example.com",
		}

		mockRepo.EXPECT().FindByID(ctx, customerID).Return(expectedCustomer, nil).Times(1)

		result, err := svc.GetCustomerByID(ctx, customerID)
		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer, result)
	})

	t.Run("Repository error is propagated", func(t *testing.T) {
		svc, mockRepo := setup()
		ctx := context.Background()
		customerID := uuid.New().String()

		expectedErr := errors.New("customer not found")
		mockRepo.EXPECT().FindByID(ctx, customerID).Return(nil, expectedErr).Times(1)

		result, err := svc.GetCustomerByID(ctx, customerID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
}

func Test_service_UpdateCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *inventory.MockClient, *salesforce.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockInventoryClient := inventory.NewMockClient(ctrl)
		mockSfClient := salesforce.NewMockClient(ctrl)
		customerUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), customerUUID.String())
		svc := &service{
			repo:            mockRepo,
			log:             zap.NewExample().Sugar(),
			sfClient:        mockSfClient,
			inventoryClient: mockInventoryClient,
		}
		return svc, ctx, mockRepo, mockInventoryClient, mockSfClient
	}

	t.Run("Successfully updates customer with Salesforce provider", func(t *testing.T) {
		svc, ctx, mockRepo, mockInventoryClient, mockSfClient := setup()

		newFirstName := "Jane"
		newLastName := "Smith"
		newEmail := "jane.smith@example.com"
		newPhoneNumber := "9876543210"
		salesforceID := "sf_user_123"

		req := &entities.UpdateCustomerRequest{
			Data: &entities.UpdateCustomerRequestBody{
				FirstName:   &newFirstName,
				LastName:    &newLastName,
				Email:       &newEmail,
				PhoneNumber: &newPhoneNumber,
			},
		}

		expectedDataToUpdate := map[string]interface{}{
			"first_name":   newFirstName,
			"last_name":    newLastName,
			"email":        newEmail,
			"phone_number": newPhoneNumber,
		}

		updatedCustomer := &entities.Customer{
			ID:           uuid.MustParse(meta.XCustomerID(ctx)),
			FirstName:    newFirstName,
			LastName:     &newLastName,
			Email:        newEmail,
			PhoneNumber:  &newPhoneNumber,
			SalesforceID: &salesforceID,
		}

		mockRepo.EXPECT().Update(ctx, expectedDataToUpdate, meta.XCustomerID(ctx)).Return(nil).Times(1)
		mockRepo.EXPECT().FindByID(ctx, meta.XCustomerID(ctx)).Return(updatedCustomer, nil).Times(1)
		mockInventoryClient.EXPECT().GetProvider().Return(providers.ProviderSalesforce).Times(1)

		// Mock the Salesforce update call that happens in goroutine
		mockSfClient.EXPECT().UpdateUserAccount(gomock.Any(), gomock.Any()).Return(nil).Times(1)

		result, err := svc.UpdateCustomer(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, updatedCustomer, result)
	})

	t.Run("No customer ID in context", func(t *testing.T) {
		svc, _, _, _, _ := setup()
		ctx := meta.WithXCustomerID(context.Background(), "")

		req := &entities.UpdateCustomerRequest{
			Data: &entities.UpdateCustomerRequestBody{
				FirstName: nullable.StringPtr("Jane"),
			},
		}

		result, err := svc.UpdateCustomer(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Contains(t, err.Error(), "Customer ID is required")
	})

	t.Run("Repository update error is propagated", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()

		req := &entities.UpdateCustomerRequest{
			Data: &entities.UpdateCustomerRequestBody{
				FirstName: nullable.StringPtr("Jane"),
			},
		}

		expectedErr := errors.New("update failed")
		mockRepo.EXPECT().Update(ctx, gomock.Any(), meta.XCustomerID(ctx)).Return(expectedErr).Times(1)

		result, err := svc.UpdateCustomer(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("Repository find error after update is propagated", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()

		req := &entities.UpdateCustomerRequest{
			Data: &entities.UpdateCustomerRequestBody{
				FirstName: nullable.StringPtr("Jane"),
			},
		}

		mockRepo.EXPECT().Update(ctx, gomock.Any(), meta.XCustomerID(ctx)).Return(nil).Times(1)

		expectedErr := errors.New("find failed")
		mockRepo.EXPECT().FindByID(ctx, meta.XCustomerID(ctx)).Return(nil, expectedErr).Times(1)

		result, err := svc.UpdateCustomer(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("Updates only provided fields with non-Salesforce provider", func(t *testing.T) {
		svc, ctx, mockRepo, mockInventoryClient, _ := setup()

		newFirstName := "Jane"

		req := &entities.UpdateCustomerRequest{
			Data: &entities.UpdateCustomerRequestBody{
				FirstName: &newFirstName,
				// Only first name provided
			},
		}

		expectedDataToUpdate := map[string]interface{}{
			"first_name": newFirstName,
		}

		updatedCustomer := &entities.Customer{
			ID:        uuid.MustParse(meta.XCustomerID(ctx)),
			FirstName: newFirstName,
		}

		mockRepo.EXPECT().Update(ctx, expectedDataToUpdate, meta.XCustomerID(ctx)).Return(nil).Times(1)
		mockRepo.EXPECT().FindByID(ctx, meta.XCustomerID(ctx)).Return(updatedCustomer, nil).Times(1)
		mockInventoryClient.EXPECT().GetProvider().Return(providers.ProviderType("other-provider")).Times(1)

		result, err := svc.UpdateCustomer(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, updatedCustomer, result)
	})
}

func Test_service_UpdateCustomerAuthorizeNetID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, mockRepo
	}

	t.Run("Successfully updates AuthorizeNet ID", func(t *testing.T) {
		svc, mockRepo := setup()
		ctx := context.Background()
		customerID := uuid.New().String()
		externalID := "auth_net_123"

		expectedDataToUpdate := map[string]interface{}{
			"authorizenet_id": externalID,
		}

		mockRepo.EXPECT().Update(ctx, expectedDataToUpdate, customerID).Return(nil).Times(1)

		err := svc.UpdateCustomerAuthorizeNetID(ctx, customerID, externalID)
		assert.NoError(t, err)
	})

	t.Run("Repository error is propagated", func(t *testing.T) {
		svc, mockRepo := setup()
		ctx := context.Background()
		customerID := uuid.New().String()
		externalID := "auth_net_123"

		expectedErr := errors.New("update failed")
		mockRepo.EXPECT().Update(ctx, gomock.Any(), customerID).Return(expectedErr).Times(1)

		err := svc.UpdateCustomerAuthorizeNetID(ctx, customerID, externalID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func Test_service_UpdateCustomerStripeID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, mockRepo
	}

	t.Run("Successfully updates Stripe ID", func(t *testing.T) {
		svc, mockRepo := setup()
		ctx := context.Background()
		customerID := uuid.New().String()
		externalID := "stripe_123"

		expectedDataToUpdate := map[string]interface{}{
			"stripe_id": externalID,
		}

		mockRepo.EXPECT().Update(ctx, expectedDataToUpdate, customerID).Return(nil).Times(1)

		err := svc.UpdateCustomerStripeID(ctx, customerID, externalID)
		assert.NoError(t, err)
	})

	t.Run("Repository error is propagated", func(t *testing.T) {
		svc, mockRepo := setup()
		ctx := context.Background()
		customerID := uuid.New().String()
		externalID := "stripe_123"

		expectedErr := errors.New("update failed")
		mockRepo.EXPECT().Update(ctx, gomock.Any(), customerID).Return(expectedErr).Times(1)

		err := svc.UpdateCustomerStripeID(ctx, customerID, externalID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func Test_service_createSalesforceUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, *repository.MockRepository, *salesforce.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockSfClient := salesforce.NewMockClient(ctrl)
		svc := &service{
			repo:     mockRepo,
			log:      zap.NewExample().Sugar(),
			sfClient: mockSfClient,
		}
		return svc, mockRepo, mockSfClient
	}

	t.Run("Successfully creates Salesforce user and updates repository", func(t *testing.T) {
		svc, mockRepo, mockSfClient := setup()
		ctx := context.Background()

		firstName := "John"
		lastName := "Doe"
		email := "john.doe@example.com"
		customerID := uuid.New().String()
		sfUserID := "sf_user_123"

		expectedSfRequest := &salesforceEntities.CreateSFUserRequest{
			FirstName:   firstName,
			LastName:    lastName,
			PersonEmail: email,
		}

		expectedSfResponse := &salesforceEntities.CreateSFUserResponse{
			ID: sfUserID,
		}

		expectedRepoUpdate := map[string]interface{}{
			"salesforce_id": sfUserID,
		}

		mockSfClient.EXPECT().CreateUserAccount(ctx, expectedSfRequest).Return(expectedSfResponse, nil).Times(1)
		mockRepo.EXPECT().Update(ctx, expectedRepoUpdate, customerID).Return(nil).Times(1)

		result, err := svc.createSalesforceUser(ctx, firstName, lastName, email, customerID)
		assert.NoError(t, err)
		assert.Equal(t, expectedSfResponse, result)
	})

	t.Run("Salesforce API error is propagated", func(t *testing.T) {
		svc, _, mockSfClient := setup()
		ctx := context.Background()

		firstName := "John"
		lastName := "Doe"
		email := "john.doe@example.com"
		customerID := uuid.New().String()

		expectedErr := errors.New("salesforce api error")
		mockSfClient.EXPECT().CreateUserAccount(ctx, gomock.Any()).Return(nil, expectedErr).Times(1)

		result, err := svc.createSalesforceUser(ctx, firstName, lastName, email, customerID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("Repository update error is propagated", func(t *testing.T) {
		svc, mockRepo, mockSfClient := setup()
		ctx := context.Background()

		firstName := "John"
		lastName := "Doe"
		email := "john.doe@example.com"
		customerID := uuid.New().String()
		sfUserID := "sf_user_123"

		expectedSfResponse := &salesforceEntities.CreateSFUserResponse{
			ID: sfUserID,
		}

		mockSfClient.EXPECT().CreateUserAccount(ctx, gomock.Any()).Return(expectedSfResponse, nil).Times(1)

		expectedErr := errors.New("repository update error")
		mockRepo.EXPECT().Update(ctx, gomock.Any(), customerID).Return(expectedErr).Times(1)

		result, err := svc.createSalesforceUser(ctx, firstName, lastName, email, customerID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
}
