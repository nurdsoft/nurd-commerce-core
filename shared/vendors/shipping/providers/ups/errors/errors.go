package errors

import (
	"encoding/json"
	"fmt"
	"github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"net/http"
)

const (
	messageFor401Error string = "Credentials are invalid or have expired. Please contact Nurdsoft Support."
)

type jsonError struct {
	Response Response `json:"response"`
}

type Response struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type UPSError struct {
	Message      string
	HttpCode     int
	ErrorCode    string
	ErrorMessage string
}

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

func (err UPSError) Error() string {
	return err.Message
}

func (err UPSError) Cause() string {
	return "ups"
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
	"UPS_RATES_ERROR":     {StatusCode: http.StatusBadRequest, Message: "Error fetching UPS shipping rates."},
	"UPS_INVALID_RATE":    {StatusCode: http.StatusInternalServerError, Message: "Invalid UPS shipping rate."},
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

// Need to get information out of this package.
func ParseError(statusCode int, responseBody []byte) (err error) {
	var upsStatusCode int
	var message string

	jsonError := jsonError{}
	err = json.Unmarshal(responseBody, &jsonError)

	if statusCode == http.StatusUnauthorized {
		// prevent frontend from being logged out
		upsStatusCode = http.StatusInternalServerError
		message = messageFor401Error
	}

	if err == nil && len(jsonError.Response.Errors) > 0 {
		return &UPSError{
			Message: jsonError.Response.Errors[0].Message,
			HttpCode:     upsStatusCode,
			ErrorCode:    jsonError.Response.Errors[0].Code,
			ErrorMessage: message,
		}
	}

	return &UPSError{
		Message:  string(responseBody),
		HttpCode: statusCode,
	}
}
