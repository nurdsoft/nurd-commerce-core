// Package metrics contains Prometheus metrics
package metrics

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http/interceptors/paths"
	prom "github.com/prometheus/client_golang/prometheus"
)

// ServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for an HTTP server.
type ServerMetrics struct {
	serverStartedCounter     *prom.CounterVec
	serverHandledCounter     *prom.CounterVec
	serverMsgReceivedCounter *prom.CounterVec
	serverMsgSentCounter     *prom.CounterVec
	serverHandledHistogram   *prom.HistogramVec

	serverContainerMemTotalGauge          prom.Gauge
	serverContainerMemUsedGauge           prom.Gauge
	serverContainerMemPercentageUsedGauge prom.Gauge
}

// NewServerMetrics returns a ServerMetrics object. Use a new instance of
// ServerMetrics when not using the default Prometheus metrics registry, for
// example when wanting to control which metrics are added to a registry as
// opposed to automatically adding metrics via init functions.
//
//nolint:dupl
func NewServerMetrics() *ServerMetrics {
	return &ServerMetrics{
		serverStartedCounter: prom.NewCounterVec(
			prom.CounterOpts{
				Namespace: "commerce_core",
				Name:      "http_server_started_total",
				Help:      "Total number of requests started on the server.",
			}, []string{"http_service", "http_method"}),
		serverHandledCounter: prom.NewCounterVec(
			prom.CounterOpts{
				Namespace: "commerce_core",
				Name:      "http_server_handled_total",
				Help:      "Total number of requests completed on the server, regardless of success or failure.",
			}, []string{"http_service", "http_method", "http_code"}),
		serverMsgReceivedCounter: prom.NewCounterVec(
			prom.CounterOpts{
				Namespace: "commerce_core",
				Name:      "http_server_msg_received_total",
				Help:      "Total number of request messages received on the server.",
			}, []string{"http_service", "http_method"}),
		serverMsgSentCounter: prom.NewCounterVec(
			prom.CounterOpts{
				Namespace: "commerce_core",
				Name:      "http_server_msg_sent_total",
				Help:      "Total number of response messages sent by the server.",
			}, []string{"http_service", "http_method"}),
		serverHandledHistogram: prom.NewHistogramVec(
			prom.HistogramOpts{
				Namespace: "commerce_core",
				Name:      "http_server_handling_seconds",
				Help:      "Histogram of response latency (seconds) of HTTP that had been application-level handled by the server.",
				Buckets:   prom.DefBuckets,
			},
			[]string{"http_service", "http_method"},
		),
		serverContainerMemTotalGauge: prom.NewGauge(
			prom.GaugeOpts{
				Namespace: "commerce_core",
				Name:      "container_memory_total",
				Help:      "Container total memory allocated",
			},
		),
		serverContainerMemUsedGauge: prom.NewGauge(
			prom.GaugeOpts{
				Namespace: "commerce_core",
				Name:      "container_memory_used",
				Help:      "Container memory used",
			},
		),
		serverContainerMemPercentageUsedGauge: prom.NewGauge(
			prom.GaugeOpts{
				Namespace: "commerce_core",
				Name:      "container_memory_used_percentage",
				Help:      "Container memory used in percentage",
			},
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (m *ServerMetrics) Describe(ch chan<- *prom.Desc) {
	m.serverStartedCounter.Describe(ch)
	m.serverHandledCounter.Describe(ch)
	m.serverMsgReceivedCounter.Describe(ch)
	m.serverMsgSentCounter.Describe(ch)
	m.serverHandledHistogram.Describe(ch)
	m.serverContainerMemTotalGauge.Describe(ch)
	m.serverContainerMemUsedGauge.Describe(ch)
	m.serverContainerMemPercentageUsedGauge.Describe(ch)
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (m *ServerMetrics) Collect(ch chan<- prom.Metric) {
	m.serverStartedCounter.Collect(ch)
	m.serverHandledCounter.Collect(ch)
	m.serverMsgReceivedCounter.Collect(ch)
	m.serverMsgSentCounter.Collect(ch)
	m.serverHandledHistogram.Collect(ch)
	m.serverContainerMemTotalGauge.Collect(ch)
	m.serverContainerMemUsedGauge.Collect(ch)
	m.serverContainerMemPercentageUsedGauge.Collect(ch)
}

type metricsHandler struct {
	router  *mux.Router
	metrics *ServerMetrics

	next http.Handler
}

func (h *metricsHandler) isValidStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}

func (h *metricsHandler) isRouteMatch(r *http.Request, match *mux.RouteMatch) bool {
	return h.router.Match(r, match)
}

func (h *metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var match mux.RouteMatch
	if !h.isRouteMatch(r, &match) {
		h.next.ServeHTTP(w, r)
		return
	}

	service, method := match.Route.GetName(), r.Method

	if paths.IsIgnoredPath(service) {
		h.next.ServeHTTP(w, r)
		return
	}

	ctx := r.Context()
	userAgent := meta.UserAgent(ctx)
	userAgentOrigin := meta.UserAgentOrigin(ctx)
	monitor := newServerReporter(h.metrics, service, method, userAgent, userAgentOrigin)
	monitor.ReceivedMessage()

	mw := interceptors.NewContentResponseWriter(w)
	h.next.ServeHTTP(mw, r)
	monitor.Handled(mw.StatusCode)

	if h.isValidStatusCode(mw.StatusCode) {
		monitor.SentMessage()
	}
}

// ServerInterceptor is an HTTP server-side interceptor that provides Prometheus monitoring for Unary requests.
func (m *ServerMetrics) ServerInterceptor(router *mux.Router, next http.Handler) http.Handler {
	containerMetricsReporter(m)

	return &metricsHandler{router, m, next}
}
