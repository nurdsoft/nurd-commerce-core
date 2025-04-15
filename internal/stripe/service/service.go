package service

import (
	"context"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/ordersclient"

	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/entities"
	salesforce "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/client"
	stripe "github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/client"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"go.uber.org/zap"
	"strings"
)

type Service interface {
	GetPaymentMethods(ctx context.Context) (*entities.GetPaymentMethodResponse, error)
	GetSetupIntent(ctx context.Context) (*entities.GetSetupIntentResponse, error)
	HandleStripeWebhook(ctx context.Context, req *entities.StripeWebhookRequest) error
}

type service struct {
	log              *zap.SugaredLogger
	config           cfg.Config
	salesforceClient salesforce.Client
	stripeClient     stripe.Client
	ordersClient     ordersclient.Client
	customerClient   customerclient.Client
}

func New(
	logger *zap.SugaredLogger,
	config cfg.Config,
	stripeClient stripe.Client,
	ordersClient ordersclient.Client,
	customerClient customerclient.Client,
) Service {
	return &service{
		log:            logger,
		config:         config,
		stripeClient:   stripeClient,
		ordersClient:   ordersClient,
		customerClient: customerClient,
	}
}

// swagger:route GET /stripe/payment-methods stripe GetPaymentMethods
//
// # Get Customer Payment Methods
// ### Get all payment methods of the customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetPaymentMethodResponse Payment methods retrieved successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetPaymentMethods(ctx context.Context) (*entities.GetPaymentMethodResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	stripeId, recentlyCreated, err := s.getCustomerStripeID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	if recentlyCreated {
		resp := &entities.GetPaymentMethodResponse{
			PaymentMethods: []entities.PaymentMethod{},
		}
		return resp, nil
	}

	result, err := s.stripeClient.GetCustomerPaymentMethods(ctx, stripeId)

	if err != nil {
		return nil, err
	}

	var paymentMethods []entities.PaymentMethod
	for _, pm := range result.PaymentMethods {
		paymentMethods = append(paymentMethods, entities.PaymentMethod{
			Id:           pm.Id,
			Brand:        pm.Brand,
			DisplayBrand: pm.DisplayBrand,
			Country:      pm.Country,
			Last4:        pm.Last4,
			ExpiryMonth:  pm.ExpiryMonth,
			ExpiryYear:   pm.ExpiryYear,
			Wallet:       pm.Wallet,
			Created:      pm.Created,
		})
	}
	resp := &entities.GetPaymentMethodResponse{
		PaymentMethods: paymentMethods,
	}

	return resp, nil
}

// swagger:route GET /stripe/setup-intent stripe GetSetupIntent
//
// # Get Setup Intent
// ### Get the setup intent for the customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetSetupIntentResponse
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetSetupIntent(ctx context.Context) (*entities.GetSetupIntentResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	stripeId, _, err := s.getCustomerStripeID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	result, err := s.stripeClient.GetSetupIntent(ctx, stripeId)

	if err != nil {
		return nil, err
	}
	setupIntent := entities.SetupIntent{
		SetupIntent:  result.SetupIntent,
		EphemeralKey: result.EphemeralKey,
		CustomerId:   result.CustomerId,
	}
	resp := &entities.GetSetupIntentResponse{
		SetupIntent: setupIntent,
	}
	return resp, nil
}

// Helper function to get the customer's stripe id or create it if it doesn't exist
// Returns the stripe id, a boolean indicating if the stripe id was created and an error
func (s *service) getCustomerStripeID(ctx context.Context, customerID string) (*string, bool, error) {
	customer, err := s.customerClient.GetCustomerByID(ctx, customerID)
	if err != nil {
		return nil, false, err
	}

	if customer.StripeId == nil {
		var fullName strings.Builder
		fullName.WriteString(customer.FirstName)
		if customer.LastName != nil {
			fullName.WriteString(" ")
			fullName.WriteString(*customer.LastName)
		}
		customerReq := &stripeEntities.CreateCustomerRequest{
			Name:  fullName.String(),
			Email: customer.Email,
		}
		stripeCustomer, err := s.stripeClient.CreateCustomer(ctx, customerReq)

		if err != nil {
			return nil, false, err
		}
		customer.StripeId = &stripeCustomer.Id

		err = s.customerClient.UpdateCustomerStripeID(ctx, customerID, stripeCustomer.Id)

		if err != nil {
			return nil, false, err
		}
		return customer.StripeId, true, nil
	}
	return customer.StripeId, false, nil
}

// swagger:route POST /stripe/webhook stripe StripeWebhookRequest
//
// Stripe Webhook
// ### Handle events from Stripe
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefaultResponse Event handled successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) HandleStripeWebhook(ctx context.Context, req *entities.StripeWebhookRequest) error {

	webhookReq := &stripeEntities.HandleWebhookEventRequest{
		Payload:   req.Payload,
		Signature: req.Signature,
	}
	event, err := s.stripeClient.GetWebhookEvent(ctx, webhookReq)
	if err != nil {
		s.log.Error("Webhook Stripe signature verification failed ", err)
		return moduleErrors.NewAPIError("STRIPE_SIGNATURE_VERIFICATION_FAILED")
	}

	switch event.Type {
	case "payment_intent.succeeded":
		s.log.Info("Payment succeeded ", "payment_intent_id ", event.ObjectId)
		err = s.ordersClient.ProcessPaymentSucceeded(ctx, event.ObjectId)
		if err != nil {
			s.log.Errorf("Error processing payment intent succeeded: %v", err)
			return nil
		}
	case "payment_intent.payment_failed":
		s.log.Info("Payment failed ", "payment_intent_id", event.ObjectId)
		err = s.ordersClient.ProcessPaymentFailed(ctx, event.ObjectId)
		if err != nil {
			s.log.Errorf("Error processing payment intent failed: %v", err)
			return nil
		}
	default:
		s.log.Warnf("Unhandled event type: %s", event.Type)
		return nil
	}

	return nil
}
