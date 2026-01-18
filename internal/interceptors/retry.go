package interceptors

import (
	"context"
	"grpc-connection-manager/pkg/logger"
	"time"

	"grpc-connection-manager/internal/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryConfig holds configuration for retry logic.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts (default: 3)
	MaxAttempts int
	// InitialBackoff is the initial backoff duration (default: 100ms)
	InitialBackoff time.Duration
	// MaxBackoff is the maximum backoff duration (default: 3s)
	MaxBackoff time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff (default: 2.0)
	BackoffMultiplier float64
	// RetryableCodes are the gRPC codes that should trigger a retry
	RetryableCodes []codes.Code
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        3 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableCodes: []codes.Code{
			codes.Unavailable,
			codes.DeadlineExceeded,
			codes.ResourceExhausted,
			codes.Aborted,
		},
	}
}

// RetryInterceptor creates a retry interceptor for gRPC unary calls.
// It automatically retries failed calls with exponential backoff.
func RetryInterceptor(cfg *RetryConfig, serviceName string, m *metrics.Metrics) grpc.UnaryClientInterceptor {
	if cfg == nil {
		cfg = DefaultRetryConfig()
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var lastErr error
		backoff := cfg.InitialBackoff

		for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
			err := invoker(ctx, method, req, reply, cc, opts...)

			if err == nil {
				if attempt > 1 {
					logger.Infof("gRPC call succeeded after %d attempts: method=%s", attempt, method)
				}
				return nil
			}

			lastErr = err
			st, ok := status.FromError(err)
			if !ok {
				return err
			}

			retryable := false
			for _, code := range cfg.RetryableCodes {
				if st.Code() == code {
					retryable = true
					break
				}
			}

			if !retryable || attempt >= cfg.MaxAttempts {
				return err
			}

			if m != nil {
				m.IncrementGRPCRetry(serviceName, method)
			}

			logger.Warnf("gRPC call failed (attempt %d/%d): method=%s, code=%s, retrying in %v",
				attempt, cfg.MaxAttempts, method, st.Code(), backoff)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}

			backoff = time.Duration(float64(backoff) * cfg.BackoffMultiplier)
			if backoff > cfg.MaxBackoff {
				backoff = cfg.MaxBackoff
			}
		}

		return lastErr
	}
}
