package service

import (
	"context"
	"encoding/json"
	"errors"

	stripeConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/config"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/errors"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/ephemeralkey"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/refund"
	"github.com/stripe/stripe-go/v81/setupintent"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.uber.org/zap"
)

type Service interface {
	CreateCustomer(ctx context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error)
	GetCustomerPaymentMethods(_ context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error)
	GetCustomerPaymentMethodById(_ context.Context, customerId, paymentMethodId *string) (*entities.GetCustomerPaymentMethodResponse, error)
	GetSetupIntent(ctx context.Context, customerId *string) (*entities.GetSetupIntentResponse, error)
	CreatePaymentIntent(ctx context.Context, req *entities.CreatePaymentIntentRequest) (*entities.CreatePaymentIntentResponse, error)
	GetWebhookEvent(_ context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error)
	Refund(_ context.Context, req *entities.RefundRequest) (*entities.RefundResponse, error)
}

func New(config stripeConfig.Config, logger *zap.SugaredLogger) (Service, error) {
	return &service{config, logger}, nil
}

type service struct {
	config stripeConfig.Config
	logger *zap.SugaredLogger
}

func (s *service) CreateCustomer(_ context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error) {
	stripe.Key = s.config.Key

	params := &stripe.CustomerParams{
		Name:  stripe.String(req.Name),
		Email: stripe.String(req.Email),
		Phone: stripe.String(req.Phone),
	}
	s.logger.Info("Creating customer in stripe-api:", req)

	result, err := customer.New(params)
	if err != nil {
		s.logger.Error("Failed to create customer in stripe-api:", err)
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) {
			s.logger.Error("Stripe error: ", stripeErr.Msg)
			switch stripeErr.Type {
			case stripe.ErrorTypeCard:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			case stripe.ErrorTypeInvalidRequest:
				return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_CREATE_USER")
			case stripe.ErrorTypeAPI:
				return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_CREATE_USER")
			}
		}
		return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_CREATE_USER")
	}

	s.logger.Info("Customer created in stripe-api ", result)

	resp := &entities.CreateCustomerResponse{
		Id: result.ID,
	}
	return resp, nil
}

func (s *service) GetCustomerPaymentMethods(_ context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error) {
	stripe.Key = s.config.Key

	params := &stripe.CustomerListPaymentMethodsParams{
		Customer: stripe.String(*customerId),
	}
	params.Limit = stripe.Int64(5)
	s.logger.Info("Fetching customer payment methods from stripe-api:", customerId)

	iter := customer.ListPaymentMethods(params)
	if iter.Err() != nil {
		s.logger.Error("Failed to fetch customer payment methods from stripe-api:", iter.Err())
		var stripeErr *stripe.Error
		if errors.As(iter.Err(), &stripeErr) {
			s.logger.Error("Stripe error: ", stripeErr.Msg)
			switch stripeErr.Type {
			case stripe.ErrorTypeInvalidRequest:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			case stripe.ErrorTypeAPI:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			}
		}
		return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_FETCH_PAYMENT_METHODS")
	}

	paymentMethods := []entities.PaymentMethod{}
	for iter.Next() {
		pm := iter.PaymentMethod()
		if pm.Card != nil {
			var wallet *string
			if pm.Card.Wallet != nil {
				walletType := string(pm.Card.Wallet.Type)
				wallet = &walletType
			}

			paymentMethods = append(paymentMethods, entities.PaymentMethod{
				Id:           pm.ID,
				Brand:        string(pm.Card.Brand),
				DisplayBrand: pm.Card.DisplayBrand,
				Country:      pm.Card.Country,
				Last4:        pm.Card.Last4,
				ExpiryMonth:  pm.Card.ExpMonth,
				ExpiryYear:   pm.Card.ExpYear,
				Wallet:       wallet,
				Created:      pm.Created,
			})
		}
	}

	resp := &entities.GetCustomerPaymentMethodsResponse{
		PaymentMethods: paymentMethods,
	}

	s.logger.Info("Customer payment methods fetched from stripe-api:", resp)
	return resp, nil
}

func (s *service) GetSetupIntent(_ context.Context, customerId *string) (*entities.GetSetupIntentResponse, error) {
	stripe.Key = s.config.Key

	ekParams := &stripe.EphemeralKeyParams{
		Customer:      stripe.String(*customerId),
		StripeVersion: stripe.String(stripe.APIVersion),
	}
	ephemeral, err := ephemeralkey.New(ekParams)
	if err != nil {
		s.logger.Error("Failed to fetch ephemeral key from stripe-api:", err)
		return nil, moduleErrors.NewAPIError("STRIPE_FAILED_TO_FETCH_EPHEMERAL_KEY")
	}
	setupIntentParams := &stripe.SetupIntentParams{
		Customer: stripe.String(*customerId),
	}

	setupIntent, err := setupintent.New(setupIntentParams)
	if err != nil {
		s.logger.Error("Failed to create a setup intent from stripe-api:", err)
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) {
			switch stripeErr.Type {
			case stripe.ErrorTypeCard:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			case stripe.ErrorTypeInvalidRequest:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			case stripe.ErrorTypeAPI:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			}
		}
		return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_CREATE_SETUP_INTENT")
	}

	resp := &entities.GetSetupIntentResponse{
		SetupIntent:  setupIntent.ClientSecret,
		EphemeralKey: ephemeral.Secret,
		CustomerId:   *customerId,
	}

	s.logger.Info("Setup intent generated successfully for customer ID:", resp.CustomerId)
	return resp, nil
}

func (s *service) CreatePaymentIntent(_ context.Context, req *entities.CreatePaymentIntentRequest) (*entities.CreatePaymentIntentResponse, error) {
	stripe.Key = s.config.Key

	integerAmount := req.Amount.Mul(decimal.NewFromInt(100)).IntPart()
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(integerAmount),
		Currency:      stripe.String(req.Currency),
		Customer:      stripe.String(*req.CustomerId),
		PaymentMethod: stripe.String(req.PaymentMethodId),
		Confirm:       stripe.Bool(true),
		OffSession:    stripe.Bool(true),
	}
	paymentIntent, err := paymentintent.New(params)
	if err != nil {
		s.logger.Error("Failed to create a payment intent from stripe-api", err)
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) {
			switch stripeErr.Code {
			case stripe.ErrorCodePaymentIntentAuthenticationFailure:
				// https://docs.stripe.com/error-codes#payment-intent-authentication-failure
				return nil, moduleErrors.NewAPIError("STRIPE_PAYMENT_INTENT_AUTHENTICATION_FAILURE")
			case stripe.ErrorCodePaymentIntentInvalidParameter:
				// https://docs.stripe.com/error-codes#payment-intent-invalid-parameter
				return nil, moduleErrors.NewAPIError("STRIPE_PAYMENT_INTENT_INVALID_PARAMETER")
			case stripe.ErrorCodePaymentIntentIncompatiblePaymentMethod:
				// https://docs.stripe.com/error-codes#payment-intent-incompatible-payment-method
				return nil, moduleErrors.NewAPIError("STRIPE_PAYMENT_INTENT_INCOMPATIBLE_PAYMENT_METHOD")
			}
		}
		return nil, moduleErrors.NewAPIError("STRIPE_PAYMENT_INTENT_ERROR")
	}

	resp := &entities.CreatePaymentIntentResponse{
		Id: paymentIntent.ID,
	}

	s.logger.Info("Payment intent created successfully:", resp)
	return resp, nil
}

func (s *service) GetWebhookEvent(_ context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error) {
	event := stripe.Event{}

	if err := json.Unmarshal(req.Payload, &event); err != nil {
		s.logger.Error("Webhook error while parsing basic request ", err)
		return nil, err
	}
	event, err := webhook.ConstructEvent(req.Payload, req.Signature, s.config.SigningSecret)
	if err != nil {
		s.logger.Error("Webhook Stripe signature verification failed ", err)
		return nil, err
	}

	if event.Type == "payment_intent.succeeded" || event.Type == "payment_intent.payment_failed" {
		var paymentIntent stripe.PaymentIntent
		err = json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			s.logger.Error("Webhook error while parsing payment intent ", err)
			return nil, err
		}
		return &entities.HandleWebhookEventResponse{
			ObjectId: paymentIntent.ID,
			Type:     string(event.Type),
		}, nil
	} else {
		s.logger.Warnf("Unhandled event type: %s", event.Type)
		return &entities.HandleWebhookEventResponse{}, nil
	}
}

// GetCustomerPaymentMethodById retrieves a specific payment method by its ID.
func (s *service) GetCustomerPaymentMethodById(_ context.Context, customerId, paymentMethodId *string) (*entities.GetCustomerPaymentMethodResponse, error) {
	stripe.Key = s.config.Key

	params := &stripe.CustomerRetrievePaymentMethodParams{
		Customer: stripe.String(*customerId),
	}

	pm, err := customer.RetrievePaymentMethod(*paymentMethodId, params)
	if err != nil {
		s.logger.Error("Failed to fetch payment method from stripe-api:", err)
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) {
			switch stripeErr.Type {
			case stripe.ErrorTypeInvalidRequest:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			case stripe.ErrorTypeAPI:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			}
		}
		return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_FETCH_PAYMENT_METHOD")
	}

	resp := &entities.GetCustomerPaymentMethodResponse{
		PaymentMethod: entities.PaymentMethod{
			Id:           pm.ID,
			Brand:        string(pm.Card.Brand),
			DisplayBrand: pm.Card.DisplayBrand,
			Country:      pm.Card.Country,
			Last4:        pm.Card.Last4,
			ExpiryMonth:  pm.Card.ExpMonth,
			ExpiryYear:   pm.Card.ExpYear,
		},
	}

	return resp, nil
}

func (s *service) Refund(_ context.Context, req *entities.RefundRequest) (*entities.RefundResponse, error) {
	stripe.Key = s.config.Key

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.PaymentIntentId),
	}

	if req.Amount.GreaterThan(decimal.Zero) {
		// Convert the amount to the smallest currency unit (e.g., cents for USD)
		integerAmount := req.Amount.Mul(decimal.NewFromInt(100)).IntPart()
		params.Amount = stripe.Int64(integerAmount)
		s.logger.Info("Refunding payment intent with amount:", integerAmount)
	} else {
		s.logger.Info("Refunding full payment intent without amount specified")
	}

	res, err := refund.New(params)
	if err != nil {
		s.logger.Error("Failed to refund payment intent from stripe-api:", err)
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) {
			switch stripeErr.Type {
			case stripe.ErrorTypeInvalidRequest:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			case stripe.ErrorTypeAPI:
				return nil, moduleErrors.NewAPIError("STRIPE_ERROR", stripeErr.Msg)
			}
		}
		return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_REFUND_PAYMENT_INTENT")
	}

	amount := decimal.NewFromInt(res.Amount).Div(decimal.NewFromInt(100))

	return &entities.RefundResponse{
		Id:       res.ID,
		Amount:   amount,
		Currency: string(res.Currency),
		Status:   string(res.Status),
		Reason:   string(res.Reason),
	}, nil
}
