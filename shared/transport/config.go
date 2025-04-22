// Package transport supports transports like HTTP, gRPC and AMQP
package transport

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	"go.uber.org/fx"
)

// Config for transport.
type Config struct {
	fx.Out

	HTTP http.Config
}
