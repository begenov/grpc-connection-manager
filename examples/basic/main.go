package main

import (
	"context"
	"grpc-connection-manager/pkg/logger"
	"time"

	"grpc-connection-manager/internal/manager"
	"grpc-connection-manager/internal/metrics"
)

func main() {
	// Create metrics instance
	m := metrics.NewMetrics()

	// Create connection manager with default config
	cfg := manager.DefaultConfig()
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
			logger.Infof("Error closing connection manager: %v", err)
		}
	}()

	ctx := context.Background()

	// Get a connection to a service
	conn, err := cm.GetConnection(ctx, "example-service", "localhost:50051")
	if err != nil {
		logger.Fatalf("Failed to get connection: %v", err)
	}

	logger.Infof("Successfully connected to service. Connection state: %s\n", conn.GetState())

	// Check health of all connections
	health := cm.HealthCheck(ctx)
	for service, status := range health {
		logger.Infof("Service: %s, State: %s, Healthy: %v\n",
			service, status.State, status.Healthy)
	}

	// Get connection count
	count := cm.GetConnectionsCount()
	logger.Infof("Total connections: %d\n", count)

	// Example: Use the connection with your gRPC client
	// client := pb.NewYourServiceClient(conn)
	// response, err := client.YourMethod(ctx, &pb.YourRequest{...})

	time.Sleep(1 * time.Second)
}
