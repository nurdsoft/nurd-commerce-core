package http

import (
	"github.com/nurdsoft/nurd-commerce-core/internal/product/endpoints"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/encode"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	goKitEndpoint "github.com/go-kit/kit/endpoint"
	goKitHTTPTransport "github.com/go-kit/kit/transport/http"
)

// RegisterTransport for http.
func RegisterTransport(
	server *httpTransport.Server,
	ep *endpoints.Endpoints,
	svcTransportClient svcTransport.Client,
) {
	registerCreateProduct(server, ep.CreateProductEndpoint, svcTransportClient)
	registerGetProduct(server, ep.GetProductEndpoint, svcTransportClient)
	registerCreateProductVariant(server, ep.CreateProductVariantEndpoint, svcTransportClient)
	registerGetProductVariant(server, ep.GetProductVariantEndpoint, svcTransportClient)
}

func registerCreateProduct(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "POST"
	path := "/product"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeCreateProductRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerGetProduct(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/product/{product_id}"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeGetProductRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}


func registerCreateProductVariant(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "POST"
	path := "/product/{product_id}/variant"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeCreateProductVariantRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerGetProductVariant(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/product/variant/{sku}"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeGetProductVariantRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}
