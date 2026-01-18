package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics collects Prometheus metrics for gRPC connections and calls.
type Metrics struct {
	// gRPC metrics
	grpcRequestsTotal       *prometheus.CounterVec
	grpcRequestDuration     *prometheus.HistogramVec
	grpcConnectionsActive   *prometheus.GaugeVec
	grpcConnectionState     *prometheus.GaugeVec
	grpcRetriesTotal        *prometheus.CounterVec
	grpcCircuitBreakerState *prometheus.GaugeVec
}

// NewMetrics creates a new Metrics instance with all Prometheus metrics initialized.
func NewMetrics() *Metrics {
	return &Metrics{
		grpcRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_client_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"service", "method", "code"},
		),
		grpcRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "grpc_client_request_duration_seconds",
				Help:    "gRPC request duration in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"service", "method"},
		),
		grpcConnectionsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "grpc_client_connections_active",
				Help: "Number of active gRPC connections",
			},
			[]string{"service"},
		),
		grpcConnectionState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "grpc_client_connection_state",
				Help: "gRPC connection state (0=Idle, 1=Connecting, 2=Ready, 3=TransientFailure, 4=Shutdown)",
			},
			[]string{"service", "state"},
		),
		grpcRetriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_client_retries_total",
				Help: "Total number of gRPC retry attempts",
			},
			[]string{"service", "method"},
		),
		grpcCircuitBreakerState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "grpc_client_circuit_breaker_state",
				Help: "Circuit breaker state (0=Closed, 1=Open, 2=HalfOpen)",
			},
			[]string{"service", "method"},
		),
	}
}
