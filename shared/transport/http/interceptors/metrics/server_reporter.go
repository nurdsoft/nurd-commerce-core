// Package metrics contains Prometheus metrics
//
//nolint:dupl
package metrics

import (
	"strconv"
	"time"
)

type serverReporter struct {
	metrics     *ServerMetrics
	serviceName string
	methodName  string
	startTime   time.Time
}

func newServerReporter(m *ServerMetrics, serviceName, methodName, userAgent, userAgentOrigin string) *serverReporter {
	r := &serverReporter{
		metrics: m,
	}
	r.startTime = time.Now()

	r.serviceName = serviceName
	r.methodName = methodName
	r.metrics.serverStartedCounter.WithLabelValues(r.serviceName, r.methodName).Inc()

	return r
}

func (r *serverReporter) ReceivedMessage() {
	r.metrics.serverMsgReceivedCounter.WithLabelValues(r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) SentMessage() {
	r.metrics.serverMsgSentCounter.WithLabelValues(r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) Handled(code int) { //nolint:interfacer
	r.metrics.serverHandledCounter.WithLabelValues(r.serviceName, r.methodName, strconv.Itoa(code)).Inc()
	r.metrics.serverHandledHistogram.WithLabelValues(r.serviceName, r.methodName).Observe(time.Since(r.startTime).Seconds())
}
