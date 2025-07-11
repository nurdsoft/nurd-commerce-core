package service

import (
	"context"

	"strings"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/ordersclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	stripeClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/client"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
	"go.uber.org/zap"
)

type Service interface {
	GetPaymentMethods(ctx context.Context) (*entities.GetPaymentMethodsResponse, error)
	GetSetupIntent(ctx context.Context) (*entities.GetSetupIntentResponse, error)
	HandleStripeWebhook(ctx context.Context, req *entities.StripeWebhookRequest) error
	GetPaymentMethod(ctx context.Context, req *entities.StripeGetPaymentMethodRequest) (*entities.GetPaymentMethodResponse, error)
}

type service struct {
	log            *zap.SugaredLogger
	stripeClient   stripeClient.Client
	ordersClient   ordersclient.Client
	customerClient customerclient.Client
}

func New(
	logger *zap.SugaredLogger,
	stripeClient stripeClient.Client,
	ordersClient ordersclient.Client,
	customerClient customerclient.Client,
) Service {
	return &service{
		log:            logger,
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
//	200: GetPaymentMethodsResponse Payment methods retrieved successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetPaymentMethods(ctx context.Context) (*entities.GetPaymentMethodsResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	stripeId, recentlyCreated, err := s.getCustomerStripeID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	if recentlyCreated {
		resp := &entities.GetPaymentMethodsResponse{
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
	resp := &entities.GetPaymentMethodsResponse{
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

// swagger:route GET /stripe/payment-method/{payment_method_id} stripe GetPaymentMethodsRequest
//
// # Get Customer Payment Method
// ### Get a specific payment method of the customer
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetPaymentMethodResponse Payment method retrieved successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetPaymentMethod(ctx context.Context, req *entities.StripeGetPaymentMethodRequest) (*entities.GetPaymentMethodResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)
	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}
	stripeId, _, err := s.getCustomerStripeID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if req.PaymentMethodId == "" {
		return nil, moduleErrors.NewAPIError("PAYMENT_METHOD_ID_REQUIRED")
	}
	paymentMethod, err := s.stripeClient.GetCustomerPaymentMethodById(ctx, stripeId, &req.PaymentMethodId)
	if err != nil {
		return nil, err
	}
	resp := &entities.GetPaymentMethodResponse{
		PaymentMethod: entities.PaymentMethod{
			Id:           paymentMethod.PaymentMethod.Id,
			Brand:        paymentMethod.PaymentMethod.Brand,
			DisplayBrand: paymentMethod.PaymentMethod.DisplayBrand,
			Country:      paymentMethod.PaymentMethod.Country,
			Last4:        paymentMethod.PaymentMethod.Last4,
			ExpiryMonth:  paymentMethod.PaymentMethod.ExpiryMonth,
			ExpiryYear:   paymentMethod.PaymentMethod.ExpiryYear,
			Wallet:       paymentMethod.PaymentMethod.Wallet,
			Created:      paymentMethod.PaymentMethod.Created,
		},
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

	if customer.StripeID == nil {
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
		customer.StripeID = &stripeCustomer.Id

		err = s.customerClient.UpdateCustomerStripeID(ctx, customerID, stripeCustomer.Id)

		if err != nil {
			return nil, false, err
		}
		return customer.StripeID, true, nil
	}
	return customer.StripeID, false, nil
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

	// TODO: handle refund.failed
	case "refund.updated":
		s.log.Info("Refund updated ", "refund_id ", event.ObjectId)
		refund, err := s.stripeClient.GetRefund(ctx, event.ObjectId)
		if err != nil {
			s.log.Errorf("Error getting refund: %v", err)
			return nil
		}

		if refund.Status == stripeEntities.StripeRefundSucceeded {
			err = s.ordersClient.ProcessRefundSucceeded(ctx, refund.Id, refund.Amount)
			if err != nil {
				s.log.Errorf("Error processing refund succeeded: %v", err)
				return nil
			}
		}
	default:
		s.log.Warnf("Unhandled event type: %s", event.Type)
		return nil
	}

	return nil
}
