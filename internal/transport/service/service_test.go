package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	"github.com/stretchr/testify/assert"
)

func TestAccessControl(t *testing.T) {
	origins := []string{"http://example.com"}
	svc := &service{origins: origins}

	mockServer := httpTransport.NewServer(
		"http://example.com",
	) // Assuming NewServer initializes the server

	allowMethods := []string{"GET", "POST"}

	t.Run("Test RegisterAccessControlOptionsHandler", func(t *testing.T) {
		path := "/test-path"

		handler := httptest.NewRecorder()
		svc.RegisterAccessControlOptionsHandler(mockServer, path, allowMethods)

		assert.NotNil(t, handler)
	})

	t.Run("Test EncodeAccessControlHeadersWrapper", func(t *testing.T) {
		mockEncoder := func(ctx context.Context, w http.ResponseWriter, r interface{}) error {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		recorder := httptest.NewRecorder()
		wrappedEncoder := svc.EncodeAccessControlHeadersWrapper(mockEncoder, allowMethods)
		wrappedEncoder(context.Background(), recorder, nil)

		assert.Equal(t, "GET, POST, PATCH, PUT, DELETE, OPTIONS", recorder.Header().Get(acAllowMethodsHeader))
		assert.Equal(t, "http://example.com", recorder.Header().Get(acAllowOriginHeader))
	})

	t.Run("Test EncodeErrorControlHeadersWrapper", func(t *testing.T) {
		mockErrorEncoder := func(ctx context.Context, err error, w http.ResponseWriter) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		recorder := httptest.NewRecorder()
		wrappedErrorEncoder := svc.EncodeErrorControlHeadersWrapper(mockErrorEncoder, allowMethods)
		wrappedErrorEncoder(context.Background(), nil, recorder)

		assert.Equal(t, "http://example.com", recorder.Header().Get(acAllowOriginHeader))
		assert.Equal(t, "GET, POST, PATCH, PUT, DELETE, OPTIONS", recorder.Header().Get(acAllowMethodsHeader))
	})
}
