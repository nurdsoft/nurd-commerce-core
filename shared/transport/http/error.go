// Package http contains http client/server with all necessary interceptor for logging, tracing, etc
package http

import (
	"encoding/json"
)

// NewError for transport
func NewError(code, message, reference string, status int) error {
	return &transportError{code, message, reference, status}
}

type jsonError struct {
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Reference string `json:"reference,omitempty"`
}

type transportError struct {
	code      string
	message   string
	reference string
	status    int
}

func (e *transportError) Error() string {
	return e.message
}

func (e *transportError) StatusCode() int {
	return e.status
}

func (e *transportError) MarshalJSON() ([]byte, error) {
	err := &jsonError{e.code, e.message, e.reference}

	return json.Marshal(err)
}
