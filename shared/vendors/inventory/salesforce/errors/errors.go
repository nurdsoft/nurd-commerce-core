package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	messageFor401Error string = "Salesforce credentials have expired. Please contact Nurdsoft Support."
)

const logPrefix = "salesforce"

type jsonError []struct {
	Message   string   `json:"message"`
	ErrorCode string   `json:"errorCode"`
	Fields    []string `json:"fields"`
}

type ErrSalesforceError struct {
	Message      string
	HttpCode     int
	ErrorCode    string
	ErrorMessage string
}

func (err ErrSalesforceError) Error() string {
	return err.Message
}

func (err ErrSalesforceError) Cause() string {
	return "salesforce"
}

// IsSalesforceError for client.
func IsSalesforceError(err error) bool {
	return SalesforceError(err) != nil
}

// SalesforceError from error
func SalesforceError(err error) *ErrSalesforceError {
	invalidResponseErr, ok := err.(*ErrSalesforceError)
	if !ok {
		return nil
	}

	return invalidResponseErr
}

// Need to get information out of this package.
func ParseSalesforceError(statusCode int, responseBody []byte) (err error) {
	var trustPortalStatusCode int
	var message string

	jsonError := jsonError{}
	err = json.Unmarshal(responseBody, &jsonError)

	if statusCode == http.StatusUnauthorized {
		// prevent frontend from being logged out
		trustPortalStatusCode = http.StatusInternalServerError
		message = messageFor401Error
	} else {
		trustPortalStatusCode = statusCode
		message = jsonError[0].Message
	}

	if err == nil {
		return &ErrSalesforceError{
			Message: fmt.Sprintf(
				logPrefix+" Error. http code: %v Error Message:  %v Error Code: %v",
				trustPortalStatusCode, jsonError[0].Message, jsonError[0].ErrorCode,
			),
			HttpCode:     trustPortalStatusCode,
			ErrorCode:    jsonError[0].ErrorCode,
			ErrorMessage: message,
		}
	}

	return &ErrSalesforceError{
		Message:  string(responseBody),
		HttpCode: statusCode,
	}
}
