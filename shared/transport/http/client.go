// Package http contains http client/server with all necessary interceptor for logging, tracing, etc
package http

import (
	"net/http"
	"time"

	retryGo "github.com/avast/retry-go"

	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/logging"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/meta"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/retry"
)

// NewClient returns new HTTP client with all interceptor like tracer, logger, metrics, etc.
func NewClient(opts ...Option) *http.Client {
	defaultOptions := options{timeout: 15 * time.Second, retries: 5}

	for _, o := range opts {
		o.apply(&defaultOptions)
	}

	var next http.RoundTripper

	if defaultOptions.transport != nil {
		next = defaultOptions.transport
	} else {
		next = http.DefaultTransport
	}

	if defaultOptions.logger != nil {
		next = logging.ClientRoundTripper(defaultOptions.logger, next)
	}

	if defaultOptions.retries > 0 {
		next = retry.ClientRoundTripper(defaultOptions.timeout, defaultOptions.retries, next, retryGo.Delay(defaultOptions.retryDelay))
	}

	next = meta.UserAgentClientRoundTripper(defaultOptions.userAgent, next)
	next = meta.RequestIDClientRoundTripper(next)

	// Wrap the RoundTripper with OTEL tracing
	// next = otelhttp.NewTransport(next)

	client := &http.Client{Transport: next}

	return client
}
