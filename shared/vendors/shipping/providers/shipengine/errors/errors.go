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

// Shipengine platform native errors
const (
	ErrInvalidToPostalCode = "Invalid to_postal_code."
	ErrEmptyToCountryCode  = "'to_country_code' should not be empty."
	ErrEmptyFromPostalCode = "'from_postal_code' should not be empty."
	ErrEmptyCarrierId      = "'carrier_id' should not be empty."
	ErrEmptyCarrierIds     = "'carrier_ids' should not be empty."
)

// Shipengine carrier errors
const (
	CarrierErrUPSMaxWeight = "UPS weight limit per package is 150 lbs."
	// Another carrier error:  "Destination country is not serviced."
	CarrierErrUPSMissingDestination = "Missing or Invalid DestinationCountry"
)

// Module-specific errors
var moduleErrors = map[string]struct {
	StatusCode int
	Message    string
}{
	"SHIPENGINE_INVALID_DELIVERY_POSTAL_CODE": {StatusCode: http.StatusBadRequest, Message: "Invalid delivery address postal code."},
	"SHIPENGINE_INVALID_ORIGIN_POSTAL_CODE":   {StatusCode: http.StatusBadRequest, Message: "Invalid origin address postal code."},
	"SHIPENGINE_MISSING_CARRIERS":             {StatusCode: http.StatusBadRequest, Message: "Missing carrier while getting shipping estimates."},
	"SHIPENGINE_ERROR_GETTING_SHIPPING_RATES": {StatusCode: http.StatusInternalServerError, Message: "Failed to get shipping estimates."},
	"SHIPENGINE_INVALID_POSTAL_CODE":          {StatusCode: http.StatusBadRequest, Message: "Invalid postal code."},
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
