// Package meta contains interceptors for meta information
package meta

import (
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/shared/meta"

	"github.com/google/uuid"
)

type requestIDHandler struct {
	next http.Handler
}

func (h *requestIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestID := h.requestID(r)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	ctx = meta.WithRequestID(ctx, requestID)

	h.next.ServeHTTP(w, r.WithContext(ctx))
}

func (h *requestIDHandler) requestID(r *http.Request) string {
	for _, header := range requestHeaders {
		requestID := r.Header.Get(header)

		if requestID != "" {
			return requestID
		}
	}

	return ""
}

// RequestIDServerHandler injects request ID into context
func RequestIDServerHandler(next http.Handler) http.Handler {
	return &requestIDHandler{next}
}

type userAgentHandler struct {
	next http.Handler
}

func (h *userAgentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userAgent := r.UserAgent()
	if userAgent != "" {
		ctx = meta.WithUserAgent(ctx, userAgent)
		ctx = meta.WithUserAgentOrigin(ctx, userAgent)
	}

	if meta.Transport(ctx) == "" {
		ctx = meta.WithTransport(ctx, "HTTP")
	}

	h.next.ServeHTTP(w, r.WithContext(ctx))
}

// UserAgentServerHandler injects user agent into context
func UserAgentServerHandler(next http.Handler) http.Handler {
	return &userAgentHandler{next}
}
