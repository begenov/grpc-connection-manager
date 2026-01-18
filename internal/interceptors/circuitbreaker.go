package interceptors

import (
	"context"
	"grpc-connection-manager/pkg/logger"
	"sync"
	"time"

	"grpc-connection-manager/internal/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CircuitBreakerState represents the state of a circuit breaker.
type CircuitBreakerState int

const (
	// StateClosed means the circuit breaker is closed and requests are allowed.
	StateClosed CircuitBreakerState = iota
	// StateOpen means the circuit breaker is open and requests are rejected.
	StateOpen
	// StateHalfOpen means the circuit breaker is testing if the service has recovered.
	StateHalfOpen
)

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of failures before opening the circuit (default: 5)
	FailureThreshold int
	// SuccessThreshold is the number of successes needed to close from half-open state (default: 2)
	SuccessThreshold int
	// Timeout is how long to wait before attempting to transition from open to half-open (default: 30s)
	Timeout time.Duration
	// RetryableCodes are the gRPC codes that should be counted as failures
	RetryableCodes []codes.Code
}

// DefaultCircuitBreakerConfig returns a CircuitBreakerConfig with sensible defaults.
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		RetryableCodes: []codes.Code{
			codes.Unavailable,
			codes.DeadlineExceeded,
			codes.ResourceExhausted,
		},
	}
}

// CircuitBreaker implements the circuit breaker pattern for gRPC calls.
type CircuitBreaker struct {
	mu          sync.Mutex
	state       CircuitBreakerState
	failures    int
	successes   int
	lastFailure time.Time
	config      *CircuitBreakerConfig
}

// NewCircuitBreaker creates a new CircuitBreaker with the given configuration.
// If cfg is nil, DefaultCircuitBreakerConfig() is used.
func NewCircuitBreaker(cfg *CircuitBreakerConfig) *CircuitBreaker {
	if cfg == nil {
		cfg = DefaultCircuitBreakerConfig()
	}
	return &CircuitBreaker{
		state:  StateClosed,
		config: cfg,
	}
}

func (cb *CircuitBreaker) Call(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	cb.mu.Lock()
	state := cb.state
	lastFailure := cb.lastFailure
	cb.mu.Unlock()

	if state == StateOpen {
		if time.Since(lastFailure) < cb.config.Timeout {
			logger.Warnf("Circuit breaker is OPEN, rejecting call: method=%s", method)
			return status.Error(codes.Unavailable, "circuit breaker is open")
		}

		cb.mu.Lock()
		if cb.state == StateOpen {
			cb.state = StateHalfOpen
			cb.successes = 0
			logger.Infof("Circuit breaker transitioning to HALF-OPEN: method=%s", method)
		}
		cb.mu.Unlock()
	}

	err := invoker(ctx, method, req, reply, cc, opts...)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		st, _ := status.FromError(err)

		retryable := false
		for _, code := range cb.config.RetryableCodes {
			if st.Code() == code {
				retryable = true
				break
			}
		}

		if retryable {
			cb.failures++
			cb.lastFailure = time.Now()

			if cb.state == StateHalfOpen {
				cb.state = StateOpen
				cb.failures = 0
				logger.Warnf("Circuit breaker transitioning to OPEN: method=%s", method)
			} else if cb.failures >= cb.config.FailureThreshold {
				cb.state = StateOpen
				logger.Warnf("Circuit breaker opened: method=%s, failures=%d", method, cb.failures)
			}
		}

		return err
	}

	cb.failures = 0

	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			logger.Infof("Circuit breaker closed: method=%s", method)
		}
	}

	return nil
}

// CircuitBreakerInterceptor creates a circuit breaker interceptor for gRPC unary calls.
// It creates a separate circuit breaker for each method to provide fine-grained control.
func CircuitBreakerInterceptor(serviceName string, cfg *CircuitBreakerConfig, m *metrics.Metrics) grpc.UnaryClientInterceptor {
	// Create a map to store circuit breakers per method
	breakers := make(map[string]*CircuitBreaker)
	var mu sync.RWMutex

	// Get or create circuit breaker for a method
	getBreaker := func(method string) *CircuitBreaker {
		mu.RLock()
		breaker, exists := breakers[method]
		mu.RUnlock()

		if exists {
			return breaker
		}

		mu.Lock()
		defer mu.Unlock()
		// Double-check after acquiring write lock
		if breaker, exists := breakers[method]; exists {
			return breaker
		}
		breaker = NewCircuitBreaker(cfg)
		breakers[method] = breaker
		return breaker
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		breaker := getBreaker(method)
		err := breaker.Call(ctx, method, req, reply, cc, invoker, opts...)

		if m != nil {
			breaker.mu.Lock()
			state := breaker.state
			breaker.mu.Unlock()
			m.UpdateGRPCCircuitBreaker(serviceName, method, int(state))
		}

		if err == nil && reply == nil {
			return status.Error(codes.Internal, "grpc reply is nil (circuit breaker interceptor bug)")
		}

		return err
	}
}
