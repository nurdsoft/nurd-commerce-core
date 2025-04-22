// Package meta contains metadata about request
package meta

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

var (
	contextKeyRequestID       = contextKey("request_id")
	contextKeyUserAgent       = contextKey("user_agent")
	contextKeyUserAgentOrigin = contextKey("user_agent_origin")
	contextKeyTransport       = contextKey("transport")
	contextKeyCustomerID      = contextKey("customer_id")
)

func (c contextKey) String() string { return string(c) }

// RequestID extracts request id from the context
func RequestID(ctx context.Context) string {
	if val, ok := ctx.Value(contextKeyRequestID).(string); ok {
		return val
	}

	return ""
}

// WithRequestID injects request id metadata to the context
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, id)
}

// UserAgent extracts user agent from the context
func UserAgent(ctx context.Context) string {
	if val, ok := ctx.Value(contextKeyUserAgent).(string); ok {
		return val
	}

	return ""
}

// WithUserAgent injects user agent metadata to the context
func WithUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, contextKeyUserAgent, ua)
}

// UserAgentOrigin extracts user agent origin from the context
func UserAgentOrigin(ctx context.Context) string {
	if val, ok := ctx.Value(contextKeyUserAgentOrigin).(string); ok {
		return val
	}

	return ""
}

// WithUserAgentOrigin injects user agent origin metadata to the context
func WithUserAgentOrigin(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, contextKeyUserAgentOrigin, ua)
}

// Transport extracts transport from the context
func Transport(ctx context.Context) string {
	if val, ok := ctx.Value(contextKeyTransport).(string); ok {
		return val
	}

	return ""
}

// WithTransport injects transport metadata to the context
func WithTransport(ctx context.Context, transport string) context.Context {
	return context.WithValue(ctx, contextKeyTransport, transport)
}

// WithXCustomerID injects customer ID metadata to the context
func WithXCustomerID(ctx context.Context, customerID string) context.Context {
	return context.WithValue(ctx, contextKeyCustomerID, customerID)
}

// XCustomerID extracts customer ID from the context
func XCustomerID(ctx context.Context) string {
	if val, ok := ctx.Value(contextKeyCustomerID).(string); ok {
		// parse customer ID from the context
		_, err := uuid.Parse(val)
		if err != nil {
			return ""
		}
		return val
	}

	return ""
}
