package http

import (
	goKitEndpoint "github.com/go-kit/kit/endpoint"
	goKitHTTPTransport "github.com/go-kit/kit/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/endpoints"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/encode"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

func RegisterTransport(
	server *httpTransport.Server,
	ep *endpoints.Endpoints,
	atc svcTransport.Client,
) {
	registerCreateOrder(server, ep.CreateOrderEndpoint, atc)
	registerListOrders(server, ep.ListOrdersEndpoint, atc)
	registerGetOrder(server, ep.GetOrderEndpoint, atc)
	registerCancelOrder(server, ep.CancelOrderEndpoint, atc)
	registerUpdateOrder(server, ep.UpdateOrderEndpoint, atc)
	registerRefundOrder(server, ep.RefundOrderEndpoint, atc)
}

func registerCreateOrder(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "POST"
	path := "/orders"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeCreateOrderRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerListOrders(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/orders"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeListOrdersRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerGetOrder(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/orders/{order_id}"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeGetOrderRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerCancelOrder(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "DELETE"
	path := "/orders/{order_id}"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeCancelOrderRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerUpdateOrder(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "PUT"
	path := "/orders/{order_reference}"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeUpdateOrderRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerRefundOrder(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "POST"
	path := "/orders/{order_reference}/refund"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeRefundOrderRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}
