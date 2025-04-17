package entities

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"net/http"
)

// Module-specific errors
var moduleErrors = map[string]struct {
	StatusCode int
	Message    string
}{
	"STRIPE_ERROR":                                      {StatusCode: http.StatusBadRequest, Message: "Stripe error."},
	"STRIPE_UNABLE_TO_CREATE_USER":                      {StatusCode: http.StatusBadRequest, Message: "Unable to create user."},
	"STRIPE_UNABLE_TO_FETCH_PAYMENT_METHODS":            {StatusCode: http.StatusInternalServerError, Message: "Unable to fetch payment methods."},
	"STRIPE_UNABLE_TO_CREATE_SETUP_INTENT":              {StatusCode: http.StatusInternalServerError, Message: "Unable to create setup intent."},
	"STRIPE_PAYMENT_INTENT_AUTHENTICATION_FAILURE":      {StatusCode: http.StatusBadRequest, Message: "Payment intent authentication failure."},
	"STRIPE_PAYMENT_INTENT_INVALID_PARAMETER":           {StatusCode: http.StatusBadRequest, Message: "Payment intent invalid parameter."},
	"STRIPE_PAYMENT_INTENT_INCOMPATIBLE_PAYMENT_METHOD": {StatusCode: http.StatusBadRequest, Message: "Payment intent incompatible payment method."},
	"STRIPE_PAYMENT_INTENT_ERROR":                       {StatusCode: http.StatusInternalServerError, Message: "Unable to create payment intent."},
	"STRIPE_FAILED_TO_FETCH_EPHEMERAL_KEY":              {StatusCode: http.StatusInternalServerError, Message: "Failed to fetch ephemeral key."},
}

func NewAPIError(errorCode string, customMessage ...string) *errors.APIError {
	if err, exists := moduleErrors[errorCode]; exists {
		message := err.Message
		if len(customMessage) > 0 {
			message = customMessage[0]
		}

		return &errors.APIError{
			ErrorCode:  errorCode, // Set dynamically
			StatusCode: err.StatusCode,
			Message:    message,
		}
	}

	// Fallback to global/common errors
	return errors.NewAPIError(errorCode, customMessage...)
}
