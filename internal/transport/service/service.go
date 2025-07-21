package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/nurdsoft/nurd-commerce-core/shared/auth"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	goKitHTTPTransport "github.com/go-kit/kit/transport/http"
)

type Service interface {
	RegisterAccessControlOptionsHandler(server *httpTransport.Server, path string, allowMethods []string)
	EncodeAccessControlHeadersWrapper(encoder goKitHTTPTransport.EncodeResponseFunc, allowMethods []string) goKitHTTPTransport.EncodeResponseFunc
	EncodeErrorControlHeadersWrapper(encoder goKitHTTPTransport.ErrorEncoder, allowMethods []string) goKitHTTPTransport.ErrorEncoder
}

// New service for transport.
func New(origins []string) Service {
	return &service{origins}
}

const (
	acAllowHeadersHeader = "Access-Control-Allow-Headers"
	acAllowMethodsHeader = "Access-Control-Allow-Methods"
	acAllowOriginHeader  = "Access-Control-Allow-Origin"
	acMaxAgeHeader       = "Access-Control-Max-Age"

	contentType = "content-type"
	//nolint:gosec
	acMaxAgeValue = "3600"
)

type service struct {
	origins []string
}

func (s *service) RegisterAccessControlOptionsHandler(server *httpTransport.Server, path string, allowMethods []string) {
	handler := goKitHTTPTransport.NewServer(
		func(_ context.Context, _ interface{}) (interface{}, error) {
			return nil, nil
		},
		func(_ context.Context, _ *http.Request) (interface{}, error) {
			return nil, nil
		},
		s.EncodeAccessControlHeadersWrapper(
			func(_ context.Context, w http.ResponseWriter, r interface{}) error {
				w.WriteHeader(http.StatusOK)
				return nil
			},
			allowMethods,
		),
		goKitHTTPTransport.ServerErrorEncoder(func(_ context.Context, err error, w http.ResponseWriter) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
	)

	server.Handle("OPTIONS", path, handler)
}

func (s *service) EncodeAccessControlHeadersWrapper(encoder goKitHTTPTransport.EncodeResponseFunc, allowMethods []string) goKitHTTPTransport.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, r interface{}) error {
		s.addAccessControlHeaders(w, allowMethods)

		return encoder(ctx, w, r)
	}
}

func (s *service) EncodeErrorControlHeadersWrapper(encoder goKitHTTPTransport.ErrorEncoder, allowMethods []string) goKitHTTPTransport.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		s.addAccessControlHeaders(w, allowMethods)

		encoder(ctx, err, w)
	}
}

func (s *service) addAccessControlHeaders(w http.ResponseWriter, allowMethods []string) {
	w.Header().Set(acAllowHeadersHeader, strings.Join([]string{
		contentType,
		acAllowOriginHeader,
		string(auth.AuthorizationKey),
		string(auth.Access),
		string(auth.CustomerIDKey),
		"Host",
		"Origin",
	}, ","))

	w.Header().Set(acAllowMethodsHeader, "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set(acAllowOriginHeader, strings.Join(s.origins, ","))
	w.Header().Set(acMaxAgeHeader, acMaxAgeValue)
	// w.Header().Set("Content-Type", "application/json")
}
