package http

import (
	goKitEndpoint "github.com/go-kit/kit/endpoint"
	goKitHTTPTransport "github.com/go-kit/kit/transport/http"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/encode"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/endpoints"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// RegisterTransport for http.
func RegisterTransport(
	server *httpTransport.Server,
	ep *endpoints.Endpoints,
	svcTransportClient svcTransport.Client,
) {

	registerAddToWishlist(server, ep.AddToWishlistEndpoint, svcTransportClient)
	registerRemoveFromWishlist(server, ep.RemoveFromWishlistEndpoint, svcTransportClient)
	registerGetWishlist(server, ep.GetWishlistEndpoint, svcTransportClient)
	registerGetMoreFromWishlist(server, ep.GetMoreFromWishlistEndpoint, svcTransportClient)
	registerGetWishlistProductTimestamps(server, ep.GetWishlistProductTimestampsEndpoint, svcTransportClient)
}

func registerAddToWishlist(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "PUT"
	path := "/wishlist"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeAddToWishlistRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerRemoveFromWishlist(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "DELETE"
	path := "/wishlist/{product_id}"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeRemoveFromWishlistRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerGetWishlist(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/wishlist"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeGetWishlistRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerGetMoreFromWishlist(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/wishlist/more"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeGetMoreFromWishlistRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerGetWishlistProductTimestamps(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "POST"
	path := "/wishlist/timestamps"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeGetWishlistProductTimestampsRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}
