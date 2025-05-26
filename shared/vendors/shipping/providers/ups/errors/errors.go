package errors

import (
	"fmt"
	"github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"net/http"
)

type ErrorObject struct {
	RequestID string   `json:"request_id,omitempty"`
	Errors    []Errors `json:"errors,omitempty"`
}

type Errors struct {
	ErrorSource string `json:"error_source,omitempty"`
	ErrorType   string `json:"error_type,omitempty"`
	ErrorCode   string `json:"error_code,omitempty"`
	Message     string `json:"message,omitempty"`
	FieldName   string `json:"field_name,omitempty"`
	FieldValue  string `json:"field_value,omitempty"`
}

func (e *ErrorObject) Error() string {
	if len(e.Errors) > 0 {
		return e.Errors[0].Message
	}
	return "unknown error"
}

func (e *ErrorObject) Unwrap() error {
	if len(e.Errors) > 0 {
		return fmt.Errorf("%s", e.Errors[0].Message)
	}
	return nil
}


// Module-specific errors
var moduleErrors = map[string]struct {
	StatusCode int
	Message    string
}{
	"UPS_INVALID_ADDRESS": {StatusCode: http.StatusBadRequest, Message: "Invalid UPS address."},
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
