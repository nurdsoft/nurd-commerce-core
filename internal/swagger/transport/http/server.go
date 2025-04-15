// Package http for conx.
package http

import (
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/internal/swagger/static"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// RegisterTransport for http.
func RegisterTransport(
	server *httpTransport.Server,
) {
	registerSwaggerUI(server)
}

func registerSwaggerUI(server *httpTransport.Server) {

	// http.FS can be used to create a http Filesystem
	var staticFS = http.FS(static.StaticFiles)
	fs := http.FileServer(staticFS)

	// sh := http.StripPrefix("/docs", fs)

	server.HandleWithPathPrefix("/docs", fs)
}
