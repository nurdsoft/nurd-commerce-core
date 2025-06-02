package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/address/errors"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/repository"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	salesforceEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
	"go.uber.org/zap"
)

type Service interface {
	CreateCustomer(ctx context.Context, req *entities.CreateCustomerRequest) (*entities.Customer, error)
	GetCustomer(ctx context.Context) (*entities.Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*entities.Customer, error)
	UpdateCustomer(ctx context.Context, req *entities.UpdateCustomerRequest) (*entities.Customer, error)
	UpdateCustomerExternalID(ctx context.Context, customerID string, externalID string, paymentProvider providers.ProviderType) error
}

type service struct {
	repo             repository.Repository
	log              *zap.SugaredLogger
	config           cfg.Config
	salesforceClient salesforce.Client
}

func New(
	repo repository.Repository,
	logger *zap.SugaredLogger,
	config cfg.Config,
	salesforceClient salesforce.Client,
) Service {
	return &service{
		repo:             repo,
		log:              logger,
		config:           config,
		salesforceClient: salesforceClient,
	}
}

// swagger:route POST /customer customers CreateCustomerRequest
//
// # Create Customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetCustomerResponse Customer created successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) CreateCustomer(ctx context.Context, req *entities.CreateCustomerRequest) (*entities.Customer, error) {
	customerID := uuid.New()
	if req.Data.ID != nil {
		customerID = *req.Data.ID
	}

	customer := &entities.Customer{
		ID:          customerID,
		Email:       req.Data.Email,
		FirstName:   req.Data.FirstName,
		LastName:    req.Data.LastName,
		PhoneNumber: req.Data.PhoneNumber,
	}

	createCustomer, err := s.repo.Create(ctx, customer)
	if err != nil {
		return nil, err
	} else {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var lastName string

			if req.Data.LastName != nil && *req.Data.LastName != "" {
				lastName = *req.Data.LastName
			} else {
				lastName = "\u200b"
			}

			_, err := s.createSalesforceUser(bgCtx, req.Data.FirstName, lastName, req.Data.Email, createCustomer.ID.String())
			if err != nil {
				s.log.Error("Error creating salesforce user account")
				return
			}
		}()
	}

	return createCustomer, nil
}

// swagger:route GET /customer customers GetCustomer
//
// # Get customer details
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetCustomerResponse
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) GetCustomer(ctx context.Context) (*entities.Customer, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	customer, err := s.repo.FindByID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// swagger:route PUT /customer customers UpdateCustomerRequest
//
// # Update Customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetCustomerResponse Customer updated successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) UpdateCustomer(ctx context.Context, req *entities.UpdateCustomerRequest) (*entities.Customer, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	err := s.repo.Update(ctx, map[string]interface{}{
		"first_name":   req.Data.FirstName,
		"last_name":    req.Data.LastName,
		"phone_number": req.Data.PhoneNumber,
	}, customerID)
	if err != nil {
		return nil, err
	}

	customer, err := s.repo.FindByID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var lastName, phone string

		if req.Data.LastName != nil && *req.Data.LastName != "" {
			lastName = *req.Data.LastName
		} else {
			// zero width space, salesforce does not accept empty strings
			lastName = "\u200b"
		}

		if req.Data.PhoneNumber != nil {
			phone = *req.Data.PhoneNumber
		}

		if customer.SalesforceID == nil {
			s.log.Info("Customer does not have a salesforce id, creating one")
			_, err := s.createSalesforceUser(bgCtx, req.Data.FirstName, lastName, customer.Email, customerID)
			if err != nil {
				s.log.Error("Error creating salesforce user account")
				return
			}
		} else {
			err = s.salesforceClient.UpdateUserAccount(bgCtx, &salesforceEntities.UpdateSFUserRequest{
				ID:        *customer.SalesforceID,
				FirstName: req.Data.FirstName,
				LastName:  lastName,
				Phone:     phone,
			})
			if err != nil {
				s.log.Error("Error updating salesforce user account")
				return
			}
		}

		s.log.Info("User details updated successfully")
	}()

	return customer, nil
}

func (s *service) GetCustomerByID(ctx context.Context, id string) (*entities.Customer, error) {
	customer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

// Create a new user in Salesforce and update the user with the Salesforce ID
func (s *service) createSalesforceUser(ctx context.Context, firstName, lastName, email, customerID string) (*salesforceEntities.CreateSFUserResponse, error) {
	res, err := s.salesforceClient.CreateUserAccount(ctx, &salesforceEntities.CreateSFUserRequest{
		FirstName:   firstName,
		LastName:    lastName,
		PersonEmail: email,
	})
	if err != nil {
		return nil, err
	}
	// update user with salesforce id
	err = s.repo.Update(ctx, map[string]interface{}{
		"salesforce_id": res.ID,
	}, customerID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *service) UpdateCustomerExternalID(ctx context.Context, customerID string, externalID string, paymentProvider providers.ProviderType) error {
	var columnName string

	switch paymentProvider {
	case providers.ProviderAuthorizeNet:
		columnName = "authorizenet_id"
	default:
		columnName = "stripe_id"
	}

	err := s.repo.Update(ctx, map[string]interface{}{
		columnName: externalID,
	}, customerID)

	if err != nil {
		return err
	}
	return nil
}
