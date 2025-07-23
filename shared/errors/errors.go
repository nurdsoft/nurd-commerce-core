package errors

import (
	"net/http"

	"github.com/pkg/errors"
)

// APIError represents a structured error response
type APIError struct {
	StatusCode int    `json:"status_code"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
}

// Implement the error interface
func (e *APIError) Error() string {
	return e.Message
}

// Common error registry (no redundant error_code)
var commonErrors = map[string]struct {
	StatusCode int
	Message    string
}{
	"DATABASE_ERROR":            {StatusCode: http.StatusInternalServerError, Message: "A database error occurred."},
	"RECORD_NOT_FOUND":          {StatusCode: http.StatusInternalServerError, Message: "Record not found."},
	"DUPLICATED_KEY":            {StatusCode: http.StatusBadRequest, Message: "This record already exists."},
	"FOREIGN_KEY_VIOLATION":     {StatusCode: http.StatusBadRequest, Message: "The referenced record does not exist or has been removed."},
	"ERROR_FETCHING_RESULTS":    {StatusCode: http.StatusInternalServerError, Message: "Error fetching results."},
	"CUSTOMER_ID_REQUIRED":      {StatusCode: http.StatusBadRequest, Message: "Customer ID is required."},
	"PRODUCT_NOT_FOUND":         {StatusCode: http.StatusNotFound, Message: "Product not found."},
	"PRODUCT_VARIANT_NOT_FOUND": {StatusCode: http.StatusNotFound, Message: "Product variant not found."},
	"VALIDATION_ERROR":          {StatusCode: http.StatusBadRequest, Message: "Validation error."},
	"INTERNAL_ERROR":            {StatusCode: http.StatusInternalServerError, Message: "An internal error occurred."},
	"PROVIDER_NOT_IMPLEMENTED":  {StatusCode: http.StatusNotImplemented, Message: "Provider not implemented."},
}

// NewAPIError dynamically retrieves messages from JSON
func NewAPIError(errorCode string, customMessage ...string) *APIError {
	if err, exists := commonErrors[errorCode]; exists {
		message := err.Message
		if len(customMessage) > 0 {
			message = customMessage[0]
		}

		return &APIError{
			ErrorCode:  errorCode, // Set error code dynamically
			StatusCode: err.StatusCode,
			Message:    message,
		}
	}

	// Default unknown error
	return &APIError{ErrorCode: "INTERNAL_ERROR", StatusCode: http.StatusInternalServerError, Message: "An internal error occurred."}
}

// IsAPIError Helper function to check error types
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
