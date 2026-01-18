package manager

import (
	"errors"
	"time"

	"google.golang.org/grpc/credentials"
)

// Config holds configuration for the ConnectionManager.
type Config struct {
	// MaxMsgSize is the maximum message size in bytes for gRPC calls (default: 1GB)
	MaxMsgSize int

	// KeepAliveTime is the interval between keepalive pings (default: 30s)
	KeepAliveTime time.Duration

	// KeepAliveTimeout is the timeout for keepalive pings (default: 5s)
	KeepAliveTimeout time.Duration

	// KeepAlivePermitWithoutStream allows keepalive pings even when there are no active streams (default: true)
	KeepAlivePermitWithoutStream bool

	// MaxReconnectDelay is the maximum delay between reconnection attempts (default: 3s)
	MaxReconnectDelay time.Duration

	// MinConnectTimeout is the minimum time to wait before attempting to reconnect (default: 10s)
	MinConnectTimeout time.Duration

	// EnableLogging enables request/response logging (default: true)
	EnableLogging bool

	// EnableMetrics enables Prometheus metrics collection (default: false)
	EnableMetrics bool

	// EnableRetry enables automatic retry on transient failures (default: true)
	EnableRetry bool

	// EnableCircuitBreaker enables circuit breaker pattern (default: true)
	EnableCircuitBreaker bool

	// TransportCredentials specifies the transport credentials to use.
	// If nil, insecure credentials are used.
	TransportCredentials credentials.TransportCredentials
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	if c.MaxMsgSize <= 0 {
		return errors.New("MaxMsgSize must be greater than 0")
	}
	if c.KeepAliveTime <= 0 {
		return errors.New("KeepAliveTime must be greater than 0")
	}
	if c.KeepAliveTimeout <= 0 {
		return errors.New("KeepAliveTimeout must be greater than 0")
	}
	if c.MaxReconnectDelay <= 0 {
		return errors.New("MaxReconnectDelay must be greater than 0")
	}
	if c.MinConnectTimeout <= 0 {
		return errors.New("MinConnectTimeout must be greater than 0")
	}
	return nil
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		MaxMsgSize:                   1024 * 1024 * 1024, // 1GB
		KeepAliveTime:                30 * time.Second,
		KeepAliveTimeout:             5 * time.Second,
		KeepAlivePermitWithoutStream: true,
		MaxReconnectDelay:            3 * time.Second,
		MinConnectTimeout:            10 * time.Second,
		EnableLogging:                true,
		EnableMetrics:                false,
		EnableRetry:                  true,
		EnableCircuitBreaker:         true,
	}
}
