package transport

import (
	"github.com/go-kit/kit/transport"
	goKitHTTPTransport "github.com/go-kit/kit/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/internal/transport/service"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	"go.uber.org/zap"
)

// Client for invoice.
type Client interface {
	RegisterAccessControlOptionsHandler(server *httpTransport.Server, path string, allowMethods []string)
	EncodeAccessControlHeadersWrapper(encoder goKitHTTPTransport.EncodeResponseFunc, allowMethods []string) goKitHTTPTransport.EncodeResponseFunc
	EncodeErrorControlHeadersWrapper(encoder goKitHTTPTransport.ErrorEncoder, allowMethods []string) goKitHTTPTransport.ErrorEncoder
	LogErrorHandler() transport.ErrorHandler
	WebSocketLogErrorHandler() transport.ErrorHandler
}

// NewClient for invoice.
func NewClient(svc service.Service, logger *zap.SugaredLogger) Client {
	return &localClient{
		svc,
		logger,
	}
}

type localClient struct {
	svc    service.Service
	logger *zap.SugaredLogger
}

func (t *localClient) WebSocketLogErrorHandler() transport.ErrorHandler {
	return LogErrorHandler(t.logger, "redesign.websocket")
}

func (t *localClient) LogErrorHandler() transport.ErrorHandler {
	return LogErrorHandler(t.logger, "redesign")
}

func (t *localClient) RegisterAccessControlOptionsHandler(server *httpTransport.Server, path string, allowMethods []string) {
	t.svc.RegisterAccessControlOptionsHandler(server, path, allowMethods)
}

func (t *localClient) EncodeAccessControlHeadersWrapper(encoder goKitHTTPTransport.EncodeResponseFunc, allowMethods []string) goKitHTTPTransport.EncodeResponseFunc {
	return t.svc.EncodeAccessControlHeadersWrapper(encoder, allowMethods)
}

func (t *localClient) EncodeErrorControlHeadersWrapper(encoder goKitHTTPTransport.ErrorEncoder, allowMethods []string) goKitHTTPTransport.ErrorEncoder {
	return t.svc.EncodeErrorControlHeadersWrapper(encoder, allowMethods)
}
