// Package http contains http client/server with all necessary interceptor for logging, tracing, etc
package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/auth"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/logging"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/maxbytesreader"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/meta"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/metrics"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/middleware"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Server defines the HTTP server
type Server struct {
	server  *http.Server
	router  *mux.Router
	options options
}

// Serve is blocking serving of HTTP requests
func (s *Server) Serve() error {
	s.registerHandlers()

	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the server from HTTP connections.
func (s *Server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.server.SetKeepAlivesEnabled(false)

	return s.server.Shutdown(ctx)
}

func (s *Server) registerHandlers() {
	var next http.Handler = s.router

	if s.options.prometheusEnabled {
		next = metrics.ServerHandler(s.router, next)
	}

	if s.options.logger != nil {
		next = logging.ServerHandler(s.router, s.options.logger, next)
	}

	next = auth.ServerHandler(next)
	next = meta.UserAgentServerHandler(next)
	next = meta.RequestIDServerHandler(next)
	next = middleware.RecoveryMiddleware(next)

	// Wrap the final handler with OTEL tracing
	next = otelhttp.NewHandler(next, "", otelhttp.WithSpanNameFormatter(
		func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		},
	))

	s.server.Handler = next
}

// HandleFunc the method and path with the handler
func (s *Server) HandleFunc(path string, handler http.Handler) {
	wrappedHandler := otelhttp.NewHandler(handler, path)
	s.router.Handle(path, wrappedHandler).Name(path)
}

// Handle the method and path with the handler
func (s *Server) Handle(method, path string, handler http.Handler) {
	wrappedHandler := otelhttp.NewHandler(handler, path)
	s.router.Handle(path, wrappedHandler).Methods(method).Name(path)
}

// HandleWithPathPrefix path with the handler
func (s *Server) HandleWithPathPrefix(path string, handler http.Handler) {
	wrappedHandler := otelhttp.NewHandler(handler, path)
	s.router.PathPrefix(path).Handler(wrappedHandler).Name(path)
}

// HandleNotFound when no route matches
func (s *Server) HandleNotFound(handler http.Handler) {
	s.router.NotFoundHandler = handler
}

// MaxFileSizeInterceptor adds http.MaxBytesReader to the handler
func (s *Server) MaxFileSizeInterceptor(maxFileSizeInMB int64, handler http.Handler) http.Handler {
	return maxbytesreader.ServerHandler(maxFileSizeInMB*1024*1024, handler)
}

// NewServer returns new HTTP server with all interceptor like tracer, logger, metrics, etc.
func NewServer(address string, opts ...Option) *Server {
	options := options{} //nolint:govet

	for _, o := range opts {
		o.apply(&options)
	}

	router := mux.NewRouter().StrictSlash(true)
	httpServer := &http.Server{
		Addr: address,
	}

	server := &Server{
		router:  router,
		server:  httpServer,
		options: options,
	}

	return server
}
