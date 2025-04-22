// Package log is based on uber zap
package log

// Config for log
type Config struct {
	FileLogEnabled bool
}

// Validate config
func (c *Config) Validate() error {
	// Empty implementation Nothing to validate
	return nil
}
