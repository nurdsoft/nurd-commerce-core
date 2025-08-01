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
	"CART_ERROR_UPDATING_CART_ITEM":     {StatusCode: http.StatusInternalServerError, Message: "Error updating cart item."},
	"CART_ITEM_NOT_FOUND":               {StatusCode: http.StatusInternalServerError, Message: "Cart item not found."},
	"CART_ERROR_GETTING_CART":           {StatusCode: http.StatusInternalServerError, Message: "Error getting cart."},
	"CART_NOT_FOUND":                    {StatusCode: http.StatusNotFound, Message: "Cart not found."},
	"CART_IS_EMPTY":                     {StatusCode: http.StatusBadRequest, Message: "Cart is empty."},
	"CART_ERROR_GETTING_CART_ITEMS":     {StatusCode: http.StatusInternalServerError, Message: "Error getting cart items."},
	"CART_ERROR_REMOVING_CART_ITEM":     {StatusCode: http.StatusInternalServerError, Message: "Error removing cart item."},
	"CART_ERROR_CLEARING_CART":          {StatusCode: http.StatusInternalServerError, Message: "Error clearing cart."},
	"CART_ERROR_UPDATING_TAX_RATE":      {StatusCode: http.StatusInternalServerError, Message: "Error updating tax rate."},
	"CART_ERROR_UPDATING_SHIPPING_RATE": {StatusCode: http.StatusInternalServerError, Message: "Error updating shipping rate."},
	"CART_NO_SHIPPING_RATES_FOUND":      {StatusCode: http.StatusInternalServerError, Message: "No shipping rates found."},
	"CART_SHIPPING_RATE_NOT_FOUND":      {StatusCode: http.StatusNotFound, Message: "Shipping rate not found."},
	"CART_ERROR_GETTING_SHIPPING_RATE":  {StatusCode: http.StatusInternalServerError, Message: "Error getting shipping rate."},
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
