package entities

import (
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/shared/errors"
)

// Module-specific errors
var moduleErrors = map[string]struct {
	StatusCode int
	Message    string
}{
	"STRIPE_SIGNATURE_VERIFICATION_FAILED": {StatusCode: http.StatusBadRequest, Message: "Stripe webhook signature verification failed"},
	"STRIPE_INVALID_REFUND_AMOUNT":         {StatusCode: http.StatusBadRequest, Message: "Invalid refund amount specified"},
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
