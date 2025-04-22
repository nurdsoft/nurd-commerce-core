// Package meta contains interceptors for meta information
package meta

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
)

type requestIDRoundTripper struct {
	next http.RoundTripper
}

func (h *requestIDRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	requestID := meta.RequestID(ctx)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	h.requestID(r, requestID)

	ctx = meta.WithRequestID(ctx, requestID)

	return h.next.RoundTrip(r.WithContext(ctx))
}

func (h *requestIDRoundTripper) requestID(r *http.Request, requestID string) {
	for _, header := range requestHeaders {
		r.Header.Set(header, requestID)
	}
}

// RequestIDClientRoundTripper injects request ID into context
func RequestIDClientRoundTripper(next http.RoundTripper) http.RoundTripper {
	return &requestIDRoundTripper{next}
}

type userAgentRoundTripper struct {
	userAgent string

	next http.RoundTripper
}

func (h *userAgentRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	previousUserAgent := meta.UserAgent(ctx)

	r.Header.Set("User-Agent", h.userAgent)
	ctx = meta.WithUserAgent(ctx, h.userAgent)

	userAgentOrigin := meta.UserAgentOrigin(ctx)
	if userAgentOrigin == "" {
		userAgentOrigin = previousUserAgent
	}

	ctx = meta.WithUserAgentOrigin(ctx, userAgentOrigin)

	return h.next.RoundTrip(r.WithContext(ctx))
}

// UserAgentClientRoundTripper injects user agent into context
func UserAgentClientRoundTripper(userAgent string, next http.RoundTripper) http.RoundTripper {
	return &userAgentRoundTripper{userAgent, next}
}
