// Package auth provides a way to interact with auth for different transports.
package auth

type headerKey string

const (
	// AuthorizationKey for bearer tokens.
	AuthorizationKey headerKey = headerKey("Authorization")
	Access           headerKey = headerKey("Access")
	CustomerIDKey    headerKey = headerKey("x-customer-id")
)
