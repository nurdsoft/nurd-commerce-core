package middleware

import (
	"encoding/json"
	httpInternal "github.com/nurdsoft/nurd-commerce-core/internal/transport/http"
	"io"
	"log"
	"net/http"
	"runtime/debug"
)

const (
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
)

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				resp := httpInternal.Response{
					Data:  &struct{}{},
					Error: &httpInternal.Error{StatusCode: http.StatusInternalServerError, ErrorCode: "BAPI_INTERNAL_ERROR", Message: "An internal error occurred."},
				}
				w.Header().Set(contentTypeHeader, jsonContentType)
				w.WriteHeader(http.StatusInternalServerError)
				_ = encodeJSONToWriter(w, resp) // nolint: errcheck
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func encodeJSONToWriter(w io.Writer, message interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(message)
}
