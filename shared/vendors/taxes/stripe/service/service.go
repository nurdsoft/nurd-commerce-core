package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/config"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/errors"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/tax/calculation"
	"go.uber.org/zap"
)

type Service interface {
	CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error)
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
	}

	// can be empty for digital goods
	if req.FromAddress != nil {
		params.ShipFromDetails = &stripe.TaxCalculationShipFromDetailsParams{
			Address: &stripe.AddressParams{
				City:       stripe.String(req.FromAddress.City),
				State:      stripe.String(req.FromAddress.State),
				PostalCode: stripe.String(req.FromAddress.PostalCode),
				Country:    stripe.String(req.FromAddress.Country),
			},
		}
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
