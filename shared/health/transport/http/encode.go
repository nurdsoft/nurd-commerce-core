// Package http contains health http transport
package http

import (
	"context"
	"encoding/json"
	"net/http"

	goKitHTTPTransport "github.com/go-kit/kit/transport/http"

	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	err = httpTransport.NewError("health_invalid", err.Error(), "", http.StatusServiceUnavailable)

	goKitHTTPTransport.DefaultErrorEncoder(ctx, err, w)
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers:", "Origin, Content-Type")

	if headerer, ok := response.(goKitHTTPTransport.Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}
	code := http.StatusOK
	if sc, ok := response.(goKitHTTPTransport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}
	return json.NewEncoder(w).Encode(response)
}
