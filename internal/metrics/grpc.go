package metrics

import (
	"time"
)

// RecordGRPCRequest records a gRPC request with its duration and status code.
func (m *Metrics) RecordGRPCRequest(service, method, code string, duration time.Duration) {
	m.grpcRequestsTotal.WithLabelValues(service, method, code).Inc()
	m.grpcRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// UpdateGRPCConnections updates the count of active gRPC connections for a service.
func (m *Metrics) UpdateGRPCConnections(service string, count int) {
	m.grpcConnectionsActive.WithLabelValues(service).Set(float64(count))
}

// UpdateGRPCConnectionState updates the connection state metric for a service.
func (m *Metrics) UpdateGRPCConnectionState(service, state string) {
	m.grpcConnectionState.WithLabelValues(service, "Idle").Set(0)
	m.grpcConnectionState.WithLabelValues(service, "Connecting").Set(0)
	m.grpcConnectionState.WithLabelValues(service, "Ready").Set(0)
	m.grpcConnectionState.WithLabelValues(service, "TransientFailure").Set(0)
	m.grpcConnectionState.WithLabelValues(service, "Shutdown").Set(0)

	m.grpcConnectionState.WithLabelValues(service, state).Set(1)
}

// IncrementGRPCRetry increments the retry counter for a gRPC method.
func (m *Metrics) IncrementGRPCRetry(service, method string) {
	m.grpcRetriesTotal.WithLabelValues(service, method).Inc()
}

// UpdateGRPCCircuitBreaker updates the circuit breaker state metric for a gRPC method.
func (m *Metrics) UpdateGRPCCircuitBreaker(service, method string, state int) {
	m.grpcCircuitBreakerState.WithLabelValues(service, method).Set(float64(state))
}
