package main

import (
	"context"
	"crypto/tls"
	"grpc-connection-manager/pkg/logger"
	"time"

	"grpc-connection-manager/internal/manager"
	"grpc-connection-manager/internal/metrics"

	"google.golang.org/grpc/credentials"
)

func main() {
	// Create metrics instance
	m := metrics.NewMetrics()

	// Create TLS credentials
	// In production, load certificates from files or use proper certificate management
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Only for testing! Use proper certificates in production
	}
	creds := credentials.NewTLS(tlsConfig)

	// Create connection manager with TLS configuration
	cfg := manager.DefaultConfig()
	cfg.TransportCredentials = creds
	cfg.EnableLogging = true
	cfg.EnableMetrics = true
	cfg.EnableRetry = true
	cfg.EnableCircuitBreaker = true

	cm, err := manager.NewConnectionManager(cfg, m)
	if err != nil {
		logger.Fatalf("Failed to create connection manager: %v", err)
	}
	defer func() {
		if err := cm.Close(); err != nil {
			logger.Errorf("Error closing connection manager: %v", err)
		}
	}()

	ctx := context.Background()

	// Get a connection to a service with TLS
	conn, err := cm.GetConnection(ctx, "secure-service", "localhost:50051")
	if err != nil {
		logger.Fatalf("Failed to get connection: %v", err)
	}

	logger.Infof("Successfully connected to secure service. Connection state: %s\n", conn.GetState())

	// Check health
	health := cm.HealthCheck(ctx)
	for service, status := range health {
		logger.Infof("Service: %s, State: %s, Healthy: %v\n",
			service, status.State, status.Healthy)
	}

	time.Sleep(1 * time.Second)
}
