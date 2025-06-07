package http

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSignature(t *testing.T) {
	// Test data
	body := []byte(`{"test":"data"}`)
	key := "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"

	h := hmac.New(sha512.New, []byte(key))
	h.Write(body)
	computedSignature := h.Sum(nil)
	sigHex := hex.EncodeToString(computedSignature)
	signature := "sha512=" + sigHex

	t.Run("valid signature", func(t *testing.T) {
		err := validateSignature(signature, body, key)
		assert.NoError(t, err)
	})

	t.Run("invalid signature", func(t *testing.T) {
		badSig := "sha512=deadbeef" + sigHex[8:]
		err := validateSignature(badSig, body, key)
		assert.Error(t, err)
	})

	t.Run("invalid signature format", func(t *testing.T) {
		badFormat := "sha256=" + sigHex
		err := validateSignature(badFormat, body, key)
		assert.Error(t, err)
	})

	t.Run("invalid key format", func(t *testing.T) {
		err := validateSignature(signature, body, "nothex")
		assert.Error(t, err)
	})

	t.Run("empty key", func(t *testing.T) {
		err := validateSignature(signature, body, "")
		assert.Error(t, err)
	})
}
