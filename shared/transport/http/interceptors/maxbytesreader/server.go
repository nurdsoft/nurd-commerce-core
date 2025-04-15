// Package maxbytesreader
package maxbytesreader

import (
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors"
)

type maxBytesReaderHandler struct {
	size int64
	next http.Handler
}

//nolint:funlen
func (h *maxBytesReaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.size)

	lw := interceptors.NewContentResponseWriter(w)
	h.next.ServeHTTP(lw, r)

}

// ServerHandler returns a new handler that adds logging.
func ServerHandler(size int64, next http.Handler) http.Handler {
	return &maxBytesReaderHandler{size, next}
}
