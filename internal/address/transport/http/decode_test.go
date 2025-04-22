package http

import (
	"bytes"
	"context"
	"github.com/nurdsoft/nurd-commerce-core/shared/nullable"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	"github.com/stretchr/testify/assert"
)

func Test_decodeAddAddressRequest(t *testing.T) {
	validBody := `{
        "full_name": "John Doe",
        "address": "123 Main St",
        "apartment": "Apt 4B",
        "city": "Anytown",
        "state_code": "CA",
        "postal_code": "12345",
        "country_code": "USA",
        "phone_number": "1234567890",
        "is_default": true
    }`

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"Valid request", validBody, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/user/address", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			got, err := decodeAddAddressRequest(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeAddAddressRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.IsType(t, &entities.AddAddressRequest{}, got)
				addAddressReq := got.(*entities.AddAddressRequest)
				assert.Equal(t, "John Doe", addAddressReq.Address.FullName)
				assert.Equal(t, "123 Main St", addAddressReq.Address.Address)
				assert.Equal(t, nullable.StringPtr("Apt 4B"), addAddressReq.Address.Apartment)
				assert.Equal(t, nullable.StringPtr("Anytown"), addAddressReq.Address.City)
				assert.Equal(t, "CA", addAddressReq.Address.StateCode)
				assert.Equal(t, "12345", addAddressReq.Address.PostalCode)
				assert.Equal(t, "USA", addAddressReq.Address.CountryCode)
				assert.Equal(t, nullable.StringPtr("1234567890"), addAddressReq.Address.PhoneNumber)
				assert.Equal(t, true, addAddressReq.Address.IsDefault)
			}
		})
	}
}

func Test_decodeGetAddressRequest(t *testing.T) {
	validID := "550e8400-e29b-41d4-a716-446655440000"
	invalidID := "invalid-id"

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"Valid ID", validID, false},
		{"Invalid ID", invalidID, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/address/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"address_id": tt.id})
			got, err := decodeGetAddressRequest(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeGetAddressRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.IsType(t, &entities.GetAddressRequest{}, got)
			}
		})
	}
}

func Test_decodeUpdateAddressRequest(t *testing.T) {
	validID := "550e8400-e29b-41d4-a716-446655440000"
	invalidID := "invalid-id"
	validBody := `{
        "full_name": "John Doe",
        "address": "123 Main St",
        "apartment": "Apt 4B",
        "city": "Anytown",
        "state_code": "CA",
        "postal_code": "12345",
        "country_code": "USA",
        "phone_number": "1234567890",
        "is_default": true
    }`
	tests := []struct {
		name    string
		id      string
		body    string
		wantErr bool
	}{
		{"Valid request", validID, validBody, false},
		{"Invalid ID", invalidID, validBody, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/address/"+tt.id, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"address_id": tt.id})
			got, err := decodeUpdateAddressRequest(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeUpdateAddressRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.IsType(t, &entities.UpdateAddressRequest{}, got)
				updateAddressReq := got.(*entities.UpdateAddressRequest)
				assert.Equal(t, "John Doe", updateAddressReq.Address.FullName)
				assert.Equal(t, "123 Main St", updateAddressReq.Address.Address)
				assert.Equal(t, nullable.StringPtr("Apt 4B"), updateAddressReq.Address.Apartment)
				assert.Equal(t, nullable.StringPtr("Anytown"), updateAddressReq.Address.City)
				assert.Equal(t, "CA", updateAddressReq.Address.StateCode)
				assert.Equal(t, "12345", updateAddressReq.Address.PostalCode)
				assert.Equal(t, "USA", updateAddressReq.Address.CountryCode)
				assert.Equal(t, nullable.StringPtr("1234567890"), updateAddressReq.Address.PhoneNumber)
				assert.Equal(t, true, updateAddressReq.Address.IsDefault)
			}
		})
	}
}

func Test_decodeDeleteAddressRequest(t *testing.T) {
	validID := "550e8400-e29b-41d4-a716-446655440000"
	invalidID := "invalid-id"

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"Valid ID", validID, false},
		{"Invalid ID", invalidID, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/address/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"address_id": tt.id})
			got, err := decodeDeleteAddressRequest(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeDeleteAddressRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.IsType(t, &entities.DeleteAddressRequest{}, got)
			}
		})
	}
}
