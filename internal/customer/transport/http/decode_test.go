package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	"github.com/stretchr/testify/assert"
)

func Test_decodeCreateCustomerRequest(t *testing.T) {
	validBody := `{"email": "customer@example.com", "first_name": "Alice"}`
	missingEmail := `{"first_name": "Alice"}`
	missingFirstName := `{"email": "customer@example.com"}`

	tests := []struct {
		name    string
		body    string
		wantErr bool
		errMsg  string
	}{
		{"Valid request", validBody, false, ""},
		{"Missing email", missingEmail, true, "Email is required"},
		{"Missing first name", missingFirstName, true, "First name is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/customer", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			decoded, err := decodeCreateCustomerRequest(context.Background(), req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, decoded)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, &entities.CreateCustomerRequest{}, decoded)
			}
		})
	}
}

func Test_decodeUpdateCustomerRequest(t *testing.T) {
	validBody := `{"email": "updated@example.com", "first_name": "Bob"}`
	validUUID := uuid.New().String()
	invalidUUID := "not-a-valid-uuid"
	tests := []struct {
		name       string
		body       string
		contextVal string
		wantErr    bool
		errMsg     string
	}{
		{"Valid request", validBody, validUUID, false, ""},
		{"Missing UUID", validBody, "", true, "Customer ID is not valid"},
		{"Invalid UUID format", validBody, invalidUUID, true, "Customer ID is not valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/customer", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.Background()
			if tt.contextVal != "" {
				ctx = sharedMeta.WithXCustomerID(ctx, tt.contextVal)
			}

			decoded, err := decodeUpdateCustomerRequest(ctx, req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, decoded)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, &entities.UpdateCustomerRequest{}, decoded)
			}
		})
	}
}
