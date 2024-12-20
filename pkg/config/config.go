package config

import "time"

var (
	// DefaultTimeout is the default timeout for AWS operations
	DefaultTimeout = 5 * time.Minute
)

// SetTimeout sets the global timeout for AWS operations
func SetTimeout(t time.Duration) {
	DefaultTimeout = t
}

// GetTimeout gets the global timeout for AWS operations
func GetTimeout() time.Duration {
	return DefaultTimeout
}
