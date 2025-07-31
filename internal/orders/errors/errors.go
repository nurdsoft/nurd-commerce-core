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
	"ORDER_NOT_FOUND":               {StatusCode: http.StatusNotFound, Message: "Order not found."},
	"ORDER_ERROR_GETTING_ITEMS":     {StatusCode: http.StatusInternalServerError, Message: "Error getting order items."},
	"ORDER_ERROR_CREATING":          {StatusCode: http.StatusInternalServerError, Message: "Error creating order."},
	"ORDER_ERROR_LISTING":           {StatusCode: http.StatusInternalServerError, Message: "Error listing orders."},
	"ORDER_ID_REQUIRED":             {StatusCode: http.StatusBadRequest, Message: "Order ID is required."},
	"ORDER_ERROR_GETTING":           {StatusCode: http.StatusInternalServerError, Message: "Error getting order."},
	"ORDER_ERROR_CANCELLING":        {StatusCode: http.StatusInternalServerError, Message: "Error cancelling order."},
	"ORDER_CANNOT_BE_CANCELLED":     {StatusCode: http.StatusInternalServerError, Message: "Order cannot be cancelled."},
	"ORDER_IS_ALREADY_CANCELLED":    {StatusCode: http.StatusNotModified, Message: "Order is already cancelled."},
	"ORDER_NOT_FOUND_BY_PAYMENT_ID": {StatusCode: http.StatusNotFound, Message: "Order not found by payment ID."},
	"ORDER_IS_NOT_PENDING":          {StatusCode: http.StatusBadRequest, Message: "Order is not pending."},
	"ORDER_REFUNDING_ERROR":         {StatusCode: http.StatusInternalServerError, Message: "Error refunding order."},
	"ORDER_INVALID_ITEMS_DATA":      {StatusCode: http.StatusBadRequest, Message: "Invalid items data in order."},
	"ORDER_INVALID_ITEM_IDENTIFIER": {StatusCode: http.StatusBadRequest, Message: "Invalid item identifier in order."},
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
