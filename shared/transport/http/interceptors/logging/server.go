// Package logging contains logs interceptors
package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nurdsoft/nurd-commerce-core/shared/log"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/paths"

	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

type loggingHandler struct {
	router *mux.Router
	logger *zap.SugaredLogger

	next http.Handler
}

func (h *loggingHandler) isRouteMatch(r *http.Request) bool {
	var match mux.RouteMatch

	return h.router.Match(r, &match)
}

//nolint:funlen
func (h *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.isRouteMatch(r) {
		h.next.ServeHTTP(w, r)
		return
	}

	service, method := r.URL.Path, r.Method
	contentType := r.Header.Get("Content-Type")

	if paths.IsIgnoredPath(service) {
		h.next.ServeHTTP(w, r)
		return
	}

	lw := interceptors.NewContentResponseWriter(w)
	startTime := time.Now()
	ctx := r.Context()

	ctxLogger := h.logger.With(
		"component", "server",
		"http_service", service,
		"http_method", method,
		"request_id", meta.RequestID(ctx),
		"user_agent", meta.UserAgent(ctx),
		"user_agent_origin", meta.UserAgentOrigin(ctx),
	)

	reqFields := []interface{}{}
	reqPayload := ""

	// read body only if content type is not multipart/form-data, since we have limited memory
	if !strings.Contains(contentType, "multipart/form-data;") {
		// TODO write a separate logger for file uploads
		if r.Body != nil && r.Body != http.NoBody {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				reqFields = append(reqFields, "http_request_read_error", err)
			}
			defer r.Body.Close()

			if json.Valid(body) {
				reqPayload = string(body)
			}

			// Restore the io.ReadCloser to its original state
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}
	}

	reqFields = append(reqFields, "http_request_content", log.Mask(reqPayload))

	ctxLogger.Infow("started call", reqFields...)

	h.next.ServeHTTP(lw, r)

	resFields := []interface{}{}

	resFields = append(resFields, "http_time_ms", durationToMilliseconds(time.Since(startTime)))

	resFields = append(resFields, "http_code", lw.StatusCode)

	resPayload := ""
	if json.Valid(lw.Data) {
		resPayload = string(lw.Data)
	}
	resFields = append(resFields, "http_response_content", log.Mask(resPayload))

	ctxLogger.Infow("finished call", resFields...)
}

// ServerHandler returns a new handler that adds logging.
func ServerHandler(router *mux.Router, logger *zap.SugaredLogger, next http.Handler) http.Handler {
	return &loggingHandler{router, logger, next}
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}
