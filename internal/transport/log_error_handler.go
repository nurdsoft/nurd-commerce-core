package transport

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/meta"

	"github.com/go-kit/kit/transport"
	"go.uber.org/zap"
)

// logErrorHandler is a server error handler implementation which logs an error.
type logErrorHandler struct {
	logger      *zap.SugaredLogger
	serviceName string
}

func (l *logErrorHandler) Handle(ctx context.Context, err error) {
	log := l.logger.With(
		"component", "server",
		"service_name", l.serviceName,
		"request_id", meta.RequestID(ctx),
		"user_agent", meta.UserAgent(ctx),
		"user_agent_origin", meta.UserAgentOrigin(ctx),
	)

	reqFields := []interface{}{}

	reqFields = append(reqFields, "error", err)

	log.Errorw("server error", reqFields...)
}

func LogErrorHandler(logger *zap.SugaredLogger, serviceName string) transport.ErrorHandler {
	return &logErrorHandler{
		logger:      logger,
		serviceName: serviceName,
	}
}
