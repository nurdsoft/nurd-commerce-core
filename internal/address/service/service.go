package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/address/errors"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/repository"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	salesforceEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
	shipengine "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/client"
	shipengineEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/entities"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"time"
)

type Service interface {
	AddAddress(ctx context.Context, req *entities.AddAddressRequest) (*entities.Address, error)
	GetAddress(ctx context.Context, req *entities.GetAddressRequest) (*entities.Address, error)
	GetAddresses(ctx context.Context) (*entities.GetAllAddressResponse, error)
	UpdateAddress(ctx context.Context, req *entities.UpdateAddressRequest) (*entities.Address, error)
	DeleteAddress(ctx context.Context, req *entities.DeleteAddressRequest) error
}

type service struct {
	repo             repository.Repository
	log              *zap.SugaredLogger
	config           cfg.Config
	shipengineClient shipengine.Client
	salesforceClient salesforce.Client
	customerClient   customerclient.Client
}

func New(
	repo repository.Repository,
	logger *zap.SugaredLogger,
	config cfg.Config,
	shipengineClient shipengine.Client,
	salesforceClient salesforce.Client,
	customerClient customerclient.Client,
) Service {
	return &service{
		repo:             repo,
		log:              logger,
		config:           config,
		shipengineClient: shipengineClient,
		salesforceClient: salesforceClient,
		customerClient:   customerClient,
	}
}

// swagger:route POST /address addresses AddAddressRequest
//
// # Add Address
// ### Add a new address for the customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	201: GetAddressResponse Address added successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) AddAddress(ctx context.Context, req *entities.AddAddressRequest) (*entities.Address, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	customer, err := uuid.Parse(customerID)
	if err != nil {
		return nil, err
	}

	err = s.validateAddress(ctx, req.Address.City, req.Address.StateCode, req.Address.PostalCode, req.Address.CountryCode)
	if err != nil {
		return nil, err
	}

	address, err := s.repo.CreateAddress(ctx, &entities.Address{
		ID:          uuid.New(),
		CustomerID:  customer,
		FullName:    req.Address.FullName,
		Address:     req.Address.Address,
		Apartment:   req.Address.Apartment,
		City:        req.Address.City,
		StateCode:   req.Address.StateCode,
		CountryCode: req.Address.CountryCode,
		PostalCode:  req.Address.PostalCode,
		PhoneNumber: req.Address.PhoneNumber,
		IsDefault:   req.Address.IsDefault,
	})
	if err != nil {
		return nil, err
	} else {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			customer, err := s.customerClient.GetCustomerByID(bgCtx, customerID)
			// TODO Handle salesforce customer creation if it doesn't exist
			if err != nil || customer.SalesforceID == nil {
				s.log.Error("Error fetching customer details")
				return
			}

			addressStreet := req.Address.Address
			if req.Address.Apartment != nil {
				addressStreet += ", " + *req.Address.Apartment
			}

			city := ""
			if req.Address.City != nil {
				city = *req.Address.City
			}

			err = s.createSalesforceUserAddress(bgCtx, customerID, address.ID.String(), &salesforceEntities.CreateSFAddressRequest{
				AccountC:               *customer.SalesforceID,
				ShippingStreetC:        addressStreet,
				ShippingCityC:          city,
				ShippingStateProvinceC: req.Address.StateCode,
				ShippingCountryC:       req.Address.CountryCode,
				ShippingZipPostalCodeC: req.Address.PostalCode,
			})
			if err != nil {
				s.log.Error("Error creating salesforce address")
				return
			}

		}()
	}

	return address, nil
}

// swagger:route GET /address/{address_id} addresses GetAddressRequest
//
// # Get Address
// ### Get a specific address of the customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetAddressResponse Address retrieved successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) GetAddress(ctx context.Context, req *entities.GetAddressRequest) (*entities.Address, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	resp, err := s.repo.GetAddress(ctx, customerID, req.AddressID.String())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// swagger:route GET /address addresses GetAddresses
//
// # Get Addresses
// ### Get all addresses of the customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetAddressResponse Addresses retrieved successfully
//	500: DefaultError Internal Server Error
func (s *service) GetAddresses(ctx context.Context) (*entities.GetAllAddressResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	resp, err := s.repo.GetAddresses(ctx, customerID)
	if err != nil {
		return nil, err
	}

	return &entities.GetAllAddressResponse{Addresses: resp}, nil
}

// swagger:route PUT /address/{address_id} addresses UpdateAddressRequest
//
// # Update Address
// ### Update an existing address of the customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetAddressResponse Address updated successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) UpdateAddress(ctx context.Context, req *entities.UpdateAddressRequest) (*entities.Address, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	customer, err := uuid.Parse(customerID)
	if err != nil {
		return nil, err
	}

	err = s.validateAddress(ctx, req.Address.City, req.Address.StateCode, req.Address.PostalCode, req.Address.CountryCode)
	if err != nil {
		return nil, err
	}

	// update address in the database
	updatedAddress, err := s.repo.UpdateAddress(ctx, &entities.Address{
		ID:          req.AddressID,
		CustomerID:  customer,
		FullName:    req.Address.FullName,
		Address:     req.Address.Address,
		Apartment:   req.Address.Apartment,
		City:        req.Address.City,
		StateCode:   req.Address.StateCode,
		PostalCode:  req.Address.PostalCode,
		CountryCode: req.Address.CountryCode,
		PhoneNumber: req.Address.PhoneNumber,
		IsDefault:   req.Address.IsDefault,
	})
	if err != nil {
		return nil, err
	} else {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			customer, err := s.customerClient.GetCustomerByID(bgCtx, customerID)
			// TODO Handle salesforce customer creation if it doesn't exist
			if err != nil || customer.SalesforceID == nil {
				s.log.Error("Error fetching customer details")
				return
			}

			addressStreet := req.Address.Address
			if req.Address.Apartment != nil {
				addressStreet += ", " + *req.Address.Apartment
			}

			city := ""
			if req.Address.City != nil {
				city = *req.Address.City
			}

			if updatedAddress.SalesforceID != nil {
				err = s.salesforceClient.UpdateUserAddress(bgCtx, &salesforceEntities.UpdateSFAddressRequest{
					AccountC:               *customer.SalesforceID,
					AddressID:              *updatedAddress.SalesforceID,
					ShippingStreetC:        addressStreet,
					ShippingCityC:          city,
					ShippingStateProvinceC: req.Address.StateCode,
					ShippingCountryC:       req.Address.CountryCode,
					ShippingZipPostalCodeC: req.Address.PostalCode,
				})
				if err != nil {
					s.log.Error("Error updating salesforce address")
					return
				}
			} else {
				// create new address in salesforce if it doesn't exist
				err := s.createSalesforceUserAddress(bgCtx, customerID, req.AddressID.String(), &salesforceEntities.CreateSFAddressRequest{
					AccountC:               *customer.SalesforceID,
					ShippingStreetC:        addressStreet,
					ShippingCityC:          city,
					ShippingStateProvinceC: req.Address.StateCode,
					ShippingCountryC:       req.Address.CountryCode,
					ShippingZipPostalCodeC: req.Address.PostalCode,
				})
				if err != nil {
					s.log.Error("Error creating salesforce address")
					return
				}
			}
		}()
	}

	return updatedAddress, nil
}

// swagger:route DELETE /address/{address_id} addresses DeleteAddressRequest
//
// # Delete Address
// ### Delete an existing address of the customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefaultResponse Address deleted successfully
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) DeleteAddress(ctx context.Context, req *entities.DeleteAddressRequest) error {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	address, err := s.repo.GetAddress(ctx, customerID, req.AddressID.String())
	if err != nil {
		return err
	}

	err = s.repo.DeleteAddress(ctx, customerID, req.AddressID.String())
	if err != nil {
		return err
	} else {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if address.SalesforceID != nil {
				err = s.salesforceClient.DeleteUserAddress(bgCtx, *address.SalesforceID)
				if err != nil {
					s.log.Error("Error deleting salesforce address")
					return
				}
			}
		}()
	}

	return nil
}

func (s *service) validateAddress(ctx context.Context, city *string, state, zipCode, country string) error {
	parsedCity := ""
	if city != nil {
		parsedCity = *city
	}

	_, err := s.shipengineClient.GetRatesEstimate(ctx,
		// From address
		shipengineEntities.ShippingAddress{
			City:    "La Vergne",
			State:   "TN",
			Zip:     "37086",
			Country: "US",
		},
		// To address

		shipengineEntities.ShippingAddress{
			City:    parsedCity,
			State:   state,
			Zip:     zipCode,
			Country: country,
		},
		// Package dimensions
		shipengineEntities.Dimensions{
			Length: decimal.NewFromFloat(1),
			Width:  decimal.NewFromFloat(1),
			Height: decimal.NewFromFloat(1),
			Weight: decimal.NewFromFloat(1),
		})
	return err
}

func (s *service) createSalesforceUserAddress(ctx context.Context, customerID, addressID string, req *salesforceEntities.CreateSFAddressRequest) error {
	res, err := s.salesforceClient.CreateUserAddress(ctx, req)
	if err != nil {
		return err
	}

	if res != nil && res.ID != "" {
		err = s.repo.UpdateAddressField(ctx, customerID, addressID, map[string]interface{}{
			"salesforce_id": res.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
