package http

import (
	goKitEndpoint "github.com/go-kit/kit/endpoint"
	goKitHTTPTransport "github.com/go-kit/kit/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/endpoints"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/encode"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// RegisterTransport for http.
func RegisterTransport(
	server *httpTransport.Server,
	ep *endpoints.Endpoints,
	svcTransportClient svcTransport.Client,
) {
	registerStripeGetPaymentMethods(server, ep.StripeGetPaymentMethodsEndpoint, svcTransportClient)
	registerStripeGetPaymentMethod(server, ep.StripeGetPaymentMethodEndpoint, svcTransportClient)
	registerStripeGetSetupIntent(server, ep.StripeGetSetupIntentEndpoint, svcTransportClient)
	registerStripeWebhook(server, ep.StripeWebhookEndpoint, svcTransportClient)
	registerStripeRefund(server, ep.StripeRefundEndpoint, svcTransportClient)
}

func registerStripeGetPaymentMethods(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/stripe/payment-methods"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeStripeGetPaymentMethods,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerStripeGetPaymentMethod(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/stripe/payment-method/{payment_method_id}"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeStripeGetPaymentMethod,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerStripeGetSetupIntent(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/stripe/setup-intent"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeStripeGetPaymentMethods,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerStripeWebhook(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "POST"
	path := "/stripe/webhook"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeStripeWebhookRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerStripeRefund(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "POST"
	path := "/stripe/refund/{payment_intent_id}"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeStripeRefundRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}
