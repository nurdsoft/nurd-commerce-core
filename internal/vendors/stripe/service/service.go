package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/config"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/errors"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/ephemeralkey"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/setupintent"
	"github.com/stripe/stripe-go/v81/tax/calculation"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.uber.org/zap"
)

type Service interface {
	CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error)
	CreateCustomer(ctx context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error)
	GetCustomerPaymentMethods(_ context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error)
	GetSetupIntent(ctx context.Context, customerId *string) (*entities.GetSetupIntentResponse, error)
	CreatePaymentIntent(ctx context.Context, req *entities.CreatePaymentIntentRequest) (*entities.CreatePaymentIntentResponse, error)
	GetWebhookEvent(_ context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error)
}

func New(config config.Config, logger *zap.SugaredLogger) (Service, error) {
	return &service{config, logger}, nil
}

type service struct {
	config config.Config
	logger *zap.SugaredLogger
}

func (s *service) CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error) {
	stripe.Key = s.config.Key

	var lineItems []*stripe.TaxCalculationLineItemParams

	for _, item := range req.TaxItems {
		lineItems = append(lineItems, &stripe.TaxCalculationLineItemParams{
			Amount:    stripe.Int64(item.Price.Mul(decimal.NewFromInt(100)).IntPart()), // Convert to minor units
			Quantity:  stripe.Int64(int64(item.Quantity)),
			TaxCode:   stripe.String(item.TaxCode),
			Reference: stripe.String(item.Reference),
		})
	}

	params := &stripe.TaxCalculationParams{
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		CustomerDetails: &stripe.TaxCalculationCustomerDetailsParams{
			Address: &stripe.AddressParams{
				Line1:      stripe.String(req.ToAddress.Line1),
				City:       stripe.String(req.ToAddress.City),
				State:      stripe.String(req.ToAddress.State),
				PostalCode: stripe.String(req.ToAddress.PostalCode),
				Country:    stripe.String(req.ToAddress.Country),
			},
			AddressSource: stripe.String(string(stripe.TaxCalculationCustomerDetailsAddressSourceShipping)),
		},
		LineItems: lineItems,
		ShippingCost: &stripe.TaxCalculationShippingCostParams{
			Amount: stripe.Int64(req.ShippingAmount.Mul(decimal.NewFromInt(100)).IntPart()), // Convert to minor units
		},
		ShipFromDetails: &stripe.TaxCalculationShipFromDetailsParams{
			Address: &stripe.AddressParams{
				City:       stripe.String(req.FromAddress.City),
				State:      stripe.String(req.FromAddress.State),
				PostalCode: stripe.String(req.FromAddress.PostalCode),
				Country:    stripe.String(req.FromAddress.Country),
			},
		},
	}

	s.logger.Info("sending tax calculation request to stripe")
	result, err := calculation.New(params)
	if err != nil {
		s.logger.Error("failed to calculate tax from stripe", err.Error())
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) {
			s.logger.Error("stripe error: ", stripeErr.Msg)
			switch stripeErr.Code {
			case stripe.ErrorCodeCustomerTaxLocationInvalid:
				// https://docs.stripe.com/error-codes#customer-tax-location-invalid
				return nil, moduleErrors.NewAPIError("STRIPE_INVALID_CUSTOMER_LOCATION")
			case stripe.ErrorCodeShippingAddressInvalid:
				// https://docs.stripe.com/error-codes#shipping-address-invalid
				return nil, moduleErrors.NewAPIError("STRIPE_INVALID_SHIPPING_ADDRESS")
			case stripe.ErrorCodeInvalidTaxLocation:
				// https://docs.stripe.com/error-codes#invalid-tax-location
				return nil, moduleErrors.NewAPIError("STRIPE_INVALID_TAX_LOCATION")
			case stripe.ErrorCodeTaxIDInvalid:
				// https://docs.stripe.com/error-codes#tax-id-invalid
				return nil, moduleErrors.NewAPIError("STRIPE_INVALID_TAX_ID")
			case stripe.ErrorCodeStripeTaxInactive:
				// https://docs.stripe.com/error-codes#stripe-tax-inactive
				return nil, moduleErrors.NewAPIError("STRIPE_TAX_IS_INACTIVE")
			case stripe.ErrorCodeTaxesCalculationFailed:
				// https://docs.stripe.com/error-codes#taxes-calculation-failed
				return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_CALCULATE_TAX")
			}
		}
		return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_CALCULATE_TAX")
	}

	taxBreakdown := result.TaxBreakdown
	var taxBreakdownJSON []byte

	if taxBreakdown != nil {
		// encode the tax breakdown to JSON
		taxBreakdownJSON, err = json.Marshal(taxBreakdown)
		if err != nil {
			s.logger.Error("failed to marshal tax breakdown to JSON", err.Error())
			return nil, moduleErrors.NewAPIError("STRIPE_UNABLE_TO_CALCULATE_TAX")
		}
	}

	return &entities.CalculateTaxResponse{
		Tax:         decimal.NewFromInt(result.TaxAmountExclusive),
		TotalAmount: decimal.NewFromInt(result.AmountTotal),
		Currency:    string(result.Currency),
		Breakdown:   taxBreakdownJSON,
	}, nil
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
		if errors.As(stripeErr, &stripeErr) {
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

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(req.Amount),
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
