package http

import (
	goKitEndpoint "github.com/go-kit/kit/endpoint"
	goKitHTTPTransport "github.com/go-kit/kit/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/internal/authorizenet/endpoints"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/encode"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// RegisterTransport for http.
func RegisterTransport(
	server *httpTransport.Server,
	ep *endpoints.Endpoints,
	svcTransportClient svcTransport.Client,
	signatureKey string,
) {
	registerGetPaymentProfiles(server, ep.GetPaymentProfilesEndpoint, svcTransportClient)
	registerCreatePaymentProfile(server, ep.CreatePaymentProfileEndpoint, svcTransportClient)
	registerWebhook(server, ep.WebhookEndpoint, svcTransportClient, signatureKey)
}

func registerGetPaymentProfiles(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "GET"
	path := "/authorizenet/payment-profiles"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeGetPaymentProfiles,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerCreatePaymentProfile(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client) {
	method := "POST"
	path := "/authorizenet/payment-profiles"

	handler := goKitHTTPTransport.NewServer(
		ep,
		decodeCreatePaymentProfileRequest,
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}

func registerWebhook(server *httpTransport.Server, ep goKitEndpoint.Endpoint, atc svcTransport.Client, signatureKey string) {
	method := "POST"
	path := "/authorizenet/webhook"

	handler := goKitHTTPTransport.NewServer(
		ep,
		NewDecodeWebhookRequest(signatureKey),
		atc.EncodeAccessControlHeadersWrapper(encode.Response, []string{method}),
		goKitHTTPTransport.ServerErrorEncoder(atc.EncodeErrorControlHeadersWrapper(encode.Error, []string{method})),
		goKitHTTPTransport.ServerErrorHandler(atc.LogErrorHandler()),
	)

	server.Handle(method, path, handler)
	atc.RegisterAccessControlOptionsHandler(server, path, []string{method})
}
