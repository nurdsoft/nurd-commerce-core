// Package metrics HTTP Prometheus monitoring interceptors for server-side HTTP.
package metrics

import (
	"net/http"

	"github.com/gorilla/mux"
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	// DefaultServerMetrics is the default instance of ServerMetrics. It is
	// intended to be used in conjunction the default Prometheus metrics
	// registry.
	DefaultServerMetrics = NewServerMetrics()
)

//nolint:gochecknoinits
func init() {
	prom.MustRegister(DefaultServerMetrics.serverStartedCounter)
	prom.MustRegister(DefaultServerMetrics.serverHandledCounter)
	prom.MustRegister(DefaultServerMetrics.serverMsgReceivedCounter)
	prom.MustRegister(DefaultServerMetrics.serverMsgSentCounter)
	prom.MustRegister(DefaultServerMetrics.serverHandledHistogram)
	prom.MustRegister(DefaultServerMetrics.serverContainerMemTotalGauge)
	prom.MustRegister(DefaultServerMetrics.serverContainerMemUsedGauge)
	prom.MustRegister(DefaultServerMetrics.serverContainerMemPercentageUsedGauge)
}

// ServerHandler is an HTTP server-side handler that provides Prometheus monitoring for requests.
func ServerHandler(router *mux.Router, next http.Handler) http.Handler {
	return DefaultServerMetrics.ServerInterceptor(router, next)
}
