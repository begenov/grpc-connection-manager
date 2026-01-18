package interceptors

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg == nil {
		t.Fatal("DefaultRetryConfig returned nil")
	}
	if cfg.MaxAttempts <= 0 {
		t.Error("MaxAttempts should be positive")
	}
	if cfg.InitialBackoff <= 0 {
		t.Error("InitialBackoff should be positive")
	}
	if cfg.MaxBackoff <= 0 {
		t.Error("MaxBackoff should be positive")
	}
	if cfg.BackoffMultiplier <= 0 {
		t.Error("BackoffMultiplier should be positive")
	}
}

func TestRetryInterceptor(t *testing.T) {
	cfg := &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		RetryableCodes:    []codes.Code{codes.Unavailable},
	}

	attempts := 0
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		if attempts < 3 {
			return status.Error(codes.Unavailable, "retry")
		}
		return nil
	}

	interceptor := RetryInterceptor(cfg, "test-service", nil)
	ctx := context.Background()
	err := interceptor(ctx, "test", nil, nil, nil, invoker)

	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}
