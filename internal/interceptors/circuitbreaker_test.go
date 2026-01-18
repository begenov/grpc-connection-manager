package interceptors

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	if cfg == nil {
		t.Fatal("DefaultCircuitBreakerConfig returned nil")
	}
	if cfg.FailureThreshold <= 0 {
		t.Error("FailureThreshold should be positive")
	}
	if cfg.SuccessThreshold <= 0 {
		t.Error("SuccessThreshold should be positive")
	}
	if cfg.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}
}

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	if cb == nil {
		t.Fatal("NewCircuitBreaker returned nil")
	}

	cfg := DefaultCircuitBreakerConfig()
	cb = NewCircuitBreaker(cfg)
	if cb == nil {
		t.Fatal("NewCircuitBreaker returned nil")
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          time.Second,
		RetryableCodes:   []codes.Code{codes.Unavailable},
	}
	cb := NewCircuitBreaker(cfg)

	// Test initial state
	if cb.state != StateClosed {
		t.Errorf("Expected initial state to be Closed, got %v", cb.state)
	}

	// Simulate failures to open circuit
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return status.Error(codes.Unavailable, "service unavailable")
	}

	ctx := context.Background()
	for i := 0; i < cfg.FailureThreshold; i++ {
		_ = cb.Call(ctx, "test", nil, nil, nil, invoker)
	}

	// Circuit should be open now
	if cb.state != StateOpen {
		t.Errorf("Expected circuit to be Open after %d failures, got %v", cfg.FailureThreshold, cb.state)
	}
}
