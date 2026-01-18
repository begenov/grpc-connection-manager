package manager

import (
	"context"

	"google.golang.org/grpc/connectivity"
)

// ConnectionHealth represents the health status of a gRPC connection.
type ConnectionHealth struct {
	State   string `json:"state"`   // Connection state (Idle, Connecting, Ready, TransientFailure, Shutdown)
	Healthy bool   `json:"healthy"` // Whether the connection is healthy
	Error   string `json:"error"`   // Error message if unhealthy
}

// HealthCheck returns the health status of all managed connections.
func (cm *ConnectionManager) HealthCheck(_ context.Context) map[string]ConnectionHealth {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]ConnectionHealth)

	for name := range cm.addresses {
		conn := cm.connections[name]
		if conn == nil {
			result[name] = ConnectionHealth{
				State:   "NotConnected",
				Healthy: false,
				Error:   "connection not established yet",
			}
			continue
		}

		state := conn.GetState()
		result[name] = ConnectionHealth{
			State:   state.String(),
			Healthy: state == connectivity.Ready,
		}

		if cm.config.EnableMetrics && cm.metrics != nil {
			cm.metrics.UpdateGRPCConnectionState(name, state.String())
		}
	}

	for name, conn := range cm.connections {
		if _, exists := cm.addresses[name]; !exists {
			state := conn.GetState()
			result[name] = ConnectionHealth{
				State:   state.String(),
				Healthy: state == connectivity.Ready,
			}
		}
	}

	return result
}
