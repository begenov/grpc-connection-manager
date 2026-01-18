package manager

import (
	"context"
	"fmt"
	"grpc-connection-manager/internal/interceptors"
	"grpc-connection-manager/internal/metrics"
	"grpc-connection-manager/pkg/logger"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ConnectionManager manages gRPC client connections with features like
// connection pooling, automatic reconnection, circuit breaking, retry logic,
// logging, and metrics collection.
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*grpc.ClientConn
	addresses   map[string]string
	config      *Config
	metrics     *metrics.Metrics
}

// NewConnectionManager creates a new ConnectionManager with the given configuration and metrics.
// If cfg is nil, DefaultConfig() is used.
// If cfg is provided, it will be validated. Returns an error if validation fails.
func NewConnectionManager(cfg *Config, m *metrics.Metrics) (*ConnectionManager, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	cm := &ConnectionManager{
		connections: make(map[string]*grpc.ClientConn),
		addresses:   make(map[string]string),
		config:      cfg,
		metrics:     m,
	}

	return cm, nil
}

// GetConnection retrieves or creates a gRPC connection for the given service.
// If address is provided, it will be used and stored for future calls.
// If address is empty, the previously stored address for the service will be used.
// Returns an error if the address is not available and connection cannot be established.
func (cm *ConnectionManager) GetConnection(ctx context.Context, serviceName string, address string) (*grpc.ClientConn, error) {
	cm.mu.Lock()
	if address != "" {
		cm.addresses[serviceName] = address
	} else {
		address = cm.addresses[serviceName]
	}
	cm.mu.Unlock()

	if address == "" {
		return nil, fmt.Errorf("address not provided and service %s not registered", serviceName)
	}

	cm.mu.RLock()
	conn := cm.connections[serviceName]
	cm.mu.RUnlock()

	if conn != nil {
		state := conn.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			return conn, nil
		}
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn = cm.connections[serviceName]; conn != nil {
		state := conn.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			return conn, nil
		}

		_ = conn.Close()
		delete(cm.connections, serviceName)
	}

	newConn, err := cm.createConnection(ctx, address, serviceName)
	if err != nil {
		logger.Warnf("Failed to create connection for %s at %s: %v (will retry on next call)", serviceName, address, err)
		return nil, fmt.Errorf("failed to create connection for %s: %w", serviceName, err)
	}

	cm.connections[serviceName] = newConn
	logger.Infof("Created gRPC connection for service: %s", serviceName)

	if cm.config.EnableMetrics && cm.metrics != nil {
		cm.metrics.UpdateGRPCConnections(serviceName, len(cm.connections))
	}

	return newConn, nil
}

func (cm *ConnectionManager) createConnection(ctx context.Context, address string, serviceName string) (*grpc.ClientConn, error) {
	creds := cm.config.TransportCredentials
	if creds == nil {
		creds = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),

		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(cm.config.MaxMsgSize),
			grpc.MaxCallSendMsgSize(cm.config.MaxMsgSize),
		),

		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cm.config.KeepAliveTime,
			Timeout:             cm.config.KeepAliveTimeout,
			PermitWithoutStream: cm.config.KeepAlivePermitWithoutStream,
		}),

		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  100 * time.Millisecond,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   cm.config.MaxReconnectDelay,
			},
			MinConnectTimeout: cm.config.MinConnectTimeout,
		}),
	}

	var unaryInterceptors []grpc.UnaryClientInterceptor

	if cm.config.EnableLogging {
		unaryInterceptors = append(unaryInterceptors,
			interceptors.LoggingInterceptor,
		)
	}

	if cm.config.EnableMetrics && cm.metrics != nil {
		unaryInterceptors = append(unaryInterceptors,
			interceptors.MetricsInterceptor(serviceName, cm.metrics),
		)
	}

	if cm.config.EnableCircuitBreaker {
		unaryInterceptors = append(unaryInterceptors,
			interceptors.CircuitBreakerInterceptor(
				serviceName,
				interceptors.DefaultCircuitBreakerConfig(),
				cm.metrics,
			),
		)
	}

	if cm.config.EnableRetry {
		unaryInterceptors = append(unaryInterceptors,
			interceptors.RetryInterceptor(
				interceptors.DefaultRetryConfig(),
				serviceName,
				cm.metrics,
			),
		)
	}

	if len(unaryInterceptors) > 0 {
		opts = append(opts,
			grpc.WithChainUnaryInterceptor(unaryInterceptors...),
		)
	}

	return grpc.DialContext(ctx, address, opts...)
}

// CloseConnection closes and removes the connection for the given service.
func (cm *ConnectionManager) CloseConnection(serviceName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn := cm.connections[serviceName]
	delete(cm.connections, serviceName)

	if cm.config.EnableMetrics && cm.metrics != nil {
		cm.metrics.UpdateGRPCConnections(serviceName, len(cm.connections))
	}

	if conn != nil {
		return conn.Close()
	}
	return nil
}

// Close closes all managed connections and cleans up resources.
func (cm *ConnectionManager) Close() error {

	cm.mu.Lock()
	defer cm.mu.Unlock()

	var lastErr error
	for name, conn := range cm.connections {
		if conn != nil {
			if err := conn.Close(); err != nil {
				logger.Errorf("Failed to close %s: %v", name, err)
				lastErr = err
			}
		}
	}
	cm.connections = make(map[string]*grpc.ClientConn)
	cm.addresses = make(map[string]string)

	if cm.config.EnableMetrics && cm.metrics != nil {
		for serviceName := range cm.connections {
			cm.metrics.UpdateGRPCConnections(serviceName, 0)
		}
	}

	return lastErr
}

// GetConnectionsCount returns the number of currently managed connections.
func (cm *ConnectionManager) GetConnectionsCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.connections)
}
