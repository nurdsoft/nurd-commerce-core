package service

import (
	"context"
	"strings"

	"github.com/nurdsoft/nurd-commerce-core/internal/authorizenet/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/ordersclient"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	authorizenetClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/client"
	authorizenetEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/entities"
	"go.uber.org/zap"
)

type Service interface {
	GetPaymentProfiles(ctx context.Context) (entities.GetPaymentProfileResponse, error)
	CreatePaymentProfile(ctx context.Context, req entities.CreatePaymentProfileRequestBody) (entities.CreatePaymentProfileResponse, error)
	HandleWebhook(ctx context.Context, req entities.WebhookRequestBody) error
}

type service struct {
	log                *zap.SugaredLogger
	config             cfg.Config
	authorizeNetClient authorizenetClient.Client
	ordersClient       ordersclient.Client
	customerClient     customerclient.Client
}

func New(
	logger *zap.SugaredLogger,
	config cfg.Config,
	authorizeNetClient authorizenetClient.Client,
	ordersClient ordersclient.Client,
	customerClient customerclient.Client,
) Service {
	return &service{
		log:                logger,
		config:             config,
		authorizeNetClient: authorizeNetClient,
		ordersClient:       ordersClient,
		customerClient:     customerClient,
	}
}

// swagger:route GET /authorizenet/payment-profiles authorizenet GetPaymentProfiles
//
// # Get Customer Payment Profiles
// ### Get all payment profiles of the customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetPaymentProfileResponse Payment profiles retrieved successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetPaymentProfiles(ctx context.Context) (entities.GetPaymentProfileResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return entities.GetPaymentProfileResponse{}, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	profileID, recentlyCreated, err := s.getProfileID(ctx, customerID)
	if err != nil {
		return entities.GetPaymentProfileResponse{}, err
	}

	if recentlyCreated {
		resp := entities.GetPaymentProfileResponse{
			PaymentProfiles: []entities.PaymentProfile{},
		}
		return resp, nil
	}

	result, err := s.authorizeNetClient.GetCustomerPaymentMethods(ctx, authorizenetEntities.GetPaymentProfilesRequest{
		ProfileID: profileID,
	})

	if err != nil {
		return entities.GetPaymentProfileResponse{}, err
	}

	paymentProfiles := make([]entities.PaymentProfile, len(result.PaymentProfiles))
	for i, pm := range result.PaymentProfiles {
		paymentProfiles[i] = entities.PaymentProfile{
			ID:             pm.ID,
			CardType:       pm.CardType,
			CardNumber:     pm.CardNumber,
			ExpirationDate: pm.ExpirationDate,
		}
	}
	resp := entities.GetPaymentProfileResponse{
		PaymentProfiles: paymentProfiles,
	}

	return resp, nil
}

// swagger:route POST /authorizenet/payment-profiles authorizenet CreatePaymentProfileRequest
//
// # Create Payment Profile
// ### Create a payment profile
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: CreatePaymentProfileResponse Customer payment profile created successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) CreatePaymentProfile(ctx context.Context, req entities.CreatePaymentProfileRequestBody) (entities.CreatePaymentProfileResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	profileID, _, err := s.getProfileID(ctx, customerID)
	if err != nil {
		return entities.CreatePaymentProfileResponse{}, err
	}

	authProfile, err := s.authorizeNetClient.CreateCustomerPaymentProfile(ctx, authorizenetEntities.CreateCustomerPaymentProfileRequest{
		ProfileID:      profileID,
		CardNumber:     req.CardNumber,
		ExpirationDate: req.ExpirationDate,
	})
	if err != nil {
		return entities.CreatePaymentProfileResponse{}, err
	}

	return entities.CreatePaymentProfileResponse{
		ProfileID:        authProfile.ProfileID,
		PaymentProfileID: authProfile.PaymentProfileID,
	}, nil
}

// Helper function to get the customer's stripe id or create it if it doesn't exist
// Returns the stripe id, a boolean indicating if the stripe id was created and an error
func (s *service) getProfileID(ctx context.Context, customerID string) (string, bool, error) {
	customer, err := s.customerClient.GetCustomerByID(ctx, customerID)
	if err != nil {
		return "", false, err
	}

	if customer.AuthorizeNetID == nil {
		var fullName strings.Builder
		fullName.WriteString(customer.FirstName)
		if customer.LastName != nil {
			fullName.WriteString(" ")
			fullName.WriteString(*customer.LastName)
		}
		customerReq := authorizenetEntities.CreateCustomerRequest{
			CustomerID:  customerID,
			Description: fullName.String(),
			Email:       customer.Email,
		}
		authProfile, err := s.authorizeNetClient.CreateCustomer(ctx, customerReq)

		if err != nil {
			return "", false, err
		}
		customer.AuthorizeNetID = &authProfile.ProfileID

		err = s.customerClient.UpdateCustomerAuthorizeNetID(ctx, customerID, authProfile.ProfileID)

		if err != nil {
			return "", false, err
		}
		return *customer.AuthorizeNetID, true, nil
	}
	return *customer.AuthorizeNetID, false, nil
}

// swagger:route POST /authorizenet/webhook authorizenet WebhookRequest
//
// Authorize.net Webhook
// ### Handle events from Authorize.net
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefaultResponse Event handled successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) HandleWebhook(ctx context.Context, req entities.WebhookRequestBody) error {
	switch req.EventType {
	case "net.authorize.payment.fraud.approved":
		s.log.Info("Payment fraud approved ", "transaction_id ", req.Payload.ID, "fraud_action", req.Payload.FraudList)
		err := s.ordersClient.ProcessPaymentSucceeded(ctx, req.Payload.ID)
		if err != nil {
			s.log.Errorf("Error processing payment succeeded: %v", err)
			return nil
		}
	case "net.authorize.payment.fraud.declined":
		s.log.Info("Payment fraud declined ", "transaction_id", req.Payload.ID, "fraud_action", req.Payload.FraudList)
		err := s.ordersClient.ProcessPaymentFailed(ctx, req.Payload.ID)
		if err != nil {
			s.log.Errorf("Error processing payment failed: %v", err)
			return nil
		}
	default:
		s.log.Warnf("Unhandled event type: %s", req.EventType)
		return nil
	}

	return nil
}
