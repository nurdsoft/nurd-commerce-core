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
	"ADDRESS_NOT_FOUND": {StatusCode: http.StatusNotFound, Message: "Address not found."},
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
