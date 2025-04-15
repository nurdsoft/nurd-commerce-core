// Package http contains health http transport
package http

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/health/endpoint"
	goKitHTTPTransport "github.com/go-kit/kit/transport/http"

	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// RegisterTransport for health
func RegisterTransport(server *httpTransport.Server, e *endpoint.Endpoints) {
	registerCheckTransport(server, e, "/health", "GET")
	registerCheckTransport(server, e, "/", "GET")
}

func registerCheckTransport(server *httpTransport.Server, e *endpoint.Endpoints, path, method string) {
	handler := goKitHTTPTransport.NewServer(
		e.CheckEndpoint,
		goKitHTTPTransport.NopRequestDecoder,
		encodeResponse,
		goKitHTTPTransport.ServerErrorEncoder(encodeError),
	)

	server.Handle(method, path, handler)
}
