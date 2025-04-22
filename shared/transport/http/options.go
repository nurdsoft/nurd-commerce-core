// Package http contains http client/server with all necessary interceptor for logging, tracing, etc
package http

import (
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// options for client
type options struct {
	openTracingTracer     opentracing.Tracer
	logger                *zap.SugaredLogger
	userAgent             string
	timeout               time.Duration
	retries               uint
	retryDelay            time.Duration
	prometheusEnabled     bool
	transport             *http.Transport
	rateLimitPerClient    rateLimit
	rateLimitPerUserAgent rateLimit
}

type rateLimit struct {
	enabled bool
	rate    time.Duration
	burst   int
}

// Option applies option
type Option interface{ apply(*options) }
type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

// WithPrometheus set enabled/disabled prometheus
func WithPrometheus(enabled bool) Option {
	return optionFunc(func(o *options) {
		o.prometheusEnabled = enabled
	})
}

// WithOpenTracingTracer set tracer
func WithOpenTracingTracer(tracer opentracing.Tracer) Option {
	return optionFunc(func(o *options) {
		o.openTracingTracer = tracer
	})
}

// WithLogger provides logging for every query
func WithLogger(logger *zap.SugaredLogger) Option {
	return optionFunc(func(o *options) {
		o.logger = logger
	})
}

// WithUserAgent for the client
func WithUserAgent(userAgent string) Option {
	return optionFunc(func(o *options) {
		o.userAgent = userAgent
	})
}

// WithTimeout for the client
func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(o *options) {
		o.timeout = timeout
	})
}

// WithRetries for the client
func WithRetries(retries uint) Option {
	return optionFunc(func(o *options) {
		o.retries = retries
	})
}

// WithRetryDelay for the client
func WithRetryDelay(delay time.Duration) Option {
	return optionFunc(func(o *options) {
		o.retryDelay = delay
	})
}

// WithTransport for the client
func WithTransport(transport *http.Transport) Option {
	return optionFunc(func(o *options) {
		o.transport = transport
	})
}

// WithClientIDRateLimit enables rate limiting for the client
// ClientID is used to determine client
func WithClientIDRateLimit(rate time.Duration, burst int) Option {
	return optionFunc(func(o *options) {
		o.rateLimitPerClient.enabled = true
		o.rateLimitPerClient.rate = rate
		o.rateLimitPerClient.burst = burst
	})
}

// WithUserAgentRateLimit enables rate limiting for the client
// UserAgent is used to determine client
func WithUserAgentRateLimit(rate time.Duration, burst int) Option {
	return optionFunc(func(o *options) {
		o.rateLimitPerUserAgent.enabled = true
		o.rateLimitPerUserAgent.rate = rate
		o.rateLimitPerUserAgent.burst = burst
	})
}
