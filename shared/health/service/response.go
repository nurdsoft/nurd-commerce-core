// Package service provides a way to do health checks
package service

// Response represents the health check response
// swagger:model HealthCheckResponse
type Response struct {
	// Message is the health status message
	// example: ok
	Message string `json:"message"`
}
