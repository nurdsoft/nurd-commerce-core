package meta

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-request-id")

	assert.Equal(t, "test-request-id", RequestID(ctx))
	assert.Equal(t, "", RequestID(context.Background()))
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-request-id")

	assert.Equal(t, "test-request-id", ctx.Value(contextKeyRequestID))
}

func TestUserAgent(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserAgent(ctx, "test-user-agent")

	assert.Equal(t, "test-user-agent", UserAgent(ctx))
	assert.Equal(t, "", UserAgent(context.Background()))
}

func TestWithUserAgent(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserAgent(ctx, "test-user-agent")

	assert.Equal(t, "test-user-agent", ctx.Value(contextKeyUserAgent))
}

func TestUserAgentOrigin(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserAgentOrigin(ctx, "test-user-agent-origin")

	assert.Equal(t, "test-user-agent-origin", UserAgentOrigin(ctx))
	assert.Equal(t, "", UserAgentOrigin(context.Background()))
}

func TestWithUserAgentOrigin(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserAgentOrigin(ctx, "test-user-agent-origin")

	assert.Equal(t, "test-user-agent-origin", ctx.Value(contextKeyUserAgentOrigin))
}

func TestTransport(t *testing.T) {
	ctx := context.Background()
	ctx = WithTransport(ctx, "test-transport")

	assert.Equal(t, "test-transport", Transport(ctx))
	assert.Equal(t, "", Transport(context.Background()))
}

func TestWithTransport(t *testing.T) {
	ctx := context.Background()
	ctx = WithTransport(ctx, "test-transport")

	assert.Equal(t, "test-transport", ctx.Value(contextKeyTransport))
}

func TestXUserId(t *testing.T) {
	ctx := context.Background()
	ctx = WithXCustomerID(ctx, "71f6d77b-c497-4bbb-b78d-9751ad0ada49")

	assert.Equal(t, "71f6d77b-c497-4bbb-b78d-9751ad0ada49", XCustomerID(ctx))
	assert.Equal(t, "", XCustomerID(context.Background()))
}

func TestWithXCustomerID(t *testing.T) {
	ctx := context.Background()
	ctx = WithXCustomerID(ctx, "test-x-customer-id")

	assert.Equal(t, "test-x-customer-id", ctx.Value(contextKeyCustomerID))
}
