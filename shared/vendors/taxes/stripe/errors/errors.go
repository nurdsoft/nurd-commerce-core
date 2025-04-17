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
	"STRIPE_INVALID_CUSTOMER_LOCATION": {StatusCode: http.StatusBadRequest, Message: "Customer tax location is invalid."},
	"STRIPE_INVALID_SHIPPING_ADDRESS":  {StatusCode: http.StatusBadRequest, Message: "Shipping address is invalid."},
	"STRIPE_INVALID_TAX_LOCATION":      {StatusCode: http.StatusBadRequest, Message: "Tax location is invalid."},
	"STRIPE_TAX_IS_INACTIVE":           {StatusCode: http.StatusBadRequest, Message: "Tax is inactive."},
	"STRIPE_UNABLE_TO_CALCULATE_TAX":   {StatusCode: http.StatusInternalServerError, Message: "Unable to calculate tax."},
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
