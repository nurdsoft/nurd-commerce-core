// Package interceptors for logging, tracing, etc
package interceptors

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// ContentResponseWriter hold data and status code of the response
type ContentResponseWriter struct {
	StatusCode int
	Data       []byte

	http.ResponseWriter
}

// NewContentResponseWriter from a http.ResponseWriter
func NewContentResponseWriter(w http.ResponseWriter) *ContentResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &ContentResponseWriter{http.StatusOK, []byte{}, w}
}

// Header returns the header map that will be sent by WriteHeader.
func (w *ContentResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// WriteHeader sends an HTTP response header with the provided status code.
func (w *ContentResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *ContentResponseWriter) Write(data []byte) (int, error) {
	w.Data = append(w.Data, data...)

	return w.ResponseWriter.Write(data)
}

// Flush sends any buffered data to the client.
func (w *ContentResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}

}

func (w *ContentResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}

	return nil, nil, errors.New("websocket: response does not implement http.Hijacker")
}
