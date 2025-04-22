package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nurdsoft/nurd-commerce-core/shared/auth"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		userIDHeader   string
		expectedUserID string
	}{
		{
			name:           "UserID header present",
			userIDHeader:   "b237f683-994a-4819-b93f-cc03b5483895",
			expectedUserID: "b237f683-994a-4819-b93f-cc03b5483895",
		},
		{
			name:           "UserID header absent",
			userIDHeader:   "",
			expectedUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := meta.XCustomerID(r.Context())
				assert.Equal(t, tt.expectedUserID, userID)
				w.WriteHeader(http.StatusOK)
			})

			handler := ServerHandler(nextHandler)

			req := httptest.NewRequest("GET", "http://example.com", nil)
			if tt.userIDHeader != "" {
				req.Header.Set(string(auth.CustomerIDKey), tt.userIDHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}
