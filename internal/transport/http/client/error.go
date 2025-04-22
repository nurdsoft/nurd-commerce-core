// Package client sending requests to network services
package client

// ErrInvalidResponse has information about what happened
type ErrInvalidResponse struct {
	HTTPStatusCode int
	Code           int
	Description    string
}

func (e *ErrInvalidResponse) Error() string {
	return e.Description
}

// IsNotFoundError for client.
func IsNotFoundError(err error) bool {
	invalidResponseErr := InvalidResponseError(err)
	if invalidResponseErr == nil {
		return false
	}

	return invalidResponseErr.Code == 404
}

// IsInvalidResponseError for client.
func IsInvalidResponseError(err error) bool {
	return InvalidResponseError(err) != nil
}

// InvalidResponseError from error
func InvalidResponseError(err error) *ErrInvalidResponse {
	invalidResponseErr, ok := err.(*ErrInvalidResponse)
	if !ok {
		return nil
	}

	return invalidResponseErr
}
