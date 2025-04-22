// Package auth contains interceptors for auth information
package auth

import (
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/shared/auth"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
)

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStr := r.Header.Get(string(auth.CustomerIDKey))
	ctx = meta.WithXCustomerID(ctx, userIDStr)

	h.next.ServeHTTP(w, r.WithContext(ctx))
}

// ServerHandler injects auth into context.
func ServerHandler(next http.Handler) http.Handler {
	return &authHandler{next}
}
