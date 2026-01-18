# gRPC Connection Manager

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A robust and feature-rich gRPC connection manager for Go applications. This library provides connection pooling, automatic reconnection, circuit breaking, retry logic, logging, and metrics collection for gRPC clients.

## Features

- 🔌 **Connection Management**: Automatic connection pooling and lifecycle management
- 🔄 **Automatic Reconnection**: Built-in reconnection logic with configurable backoff
- 🛡️ **Circuit Breaker**: Protect your services from cascading failures
- 🔁 **Retry Logic**: Configurable retry mechanism with exponential backoff
- 📊 **Metrics**: Prometheus metrics integration for monitoring
- 📝 **Logging**: Structured logging with zap logger
- 🏥 **Health Checks**: Built-in health check functionality
- ⚙️ **Configurable**: Highly configurable with sensible defaults

## Installation

```bash
go get github.com/yourusername/grpc-connection-manager
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "grpc-connection-manager/internal/manager"
    "grpc-connection-manager/internal/metrics"
    "google.golang.org/grpc"
)

func main() {
    // Create metrics instance
    m := metrics.NewMetrics()
    
    // Create connection manager with default config
    cfg := manager.DefaultConfig()
    cm, err := manager.NewConnectionManager(cfg, m)
    if err != nil {
        log.Fatal(err)
    }
    defer cm.Close()
    
    // Get a connection
    ctx := context.Background()
    conn, err := cm.GetConnection(ctx, "my-service", "localhost:50051")
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the connection with your gRPC client
    // client := pb.NewYourServiceClient(conn)
    // ...
}
```

## Configuration

You can customize the connection manager behavior:

```go
cfg := &manager.Config{
    MaxMsgSize:                   4 * 1024 * 1024, // 4MB
    KeepAliveTime:                30 * time.Second,
    KeepAliveTimeout:             5 * time.Second,
    KeepAlivePermitWithoutStream: true,
    MaxReconnectDelay:            3 * time.Second,
    MinConnectTimeout:            10 * time.Second,
    EnableLogging:                true,
    EnableMetrics:                true,
    EnableRetry:                  true,
    EnableCircuitBreaker:         true,
}

cm, err := manager.NewConnectionManager(cfg, metrics.NewMetrics())
if err != nil {
    log.Fatal(err)
}
```

### TLS/SSL Support

The connection manager supports TLS/SSL connections:

```go
import (
    "crypto/tls"
    "google.golang.org/grpc/credentials"
)

// Create TLS credentials
tlsConfig := &tls.Config{
    ServerName: "your-server-name",
    // Load certificates from files in production
}
creds := credentials.NewTLS(tlsConfig)

cfg := manager.DefaultConfig()
cfg.TransportCredentials = creds

cm, err := manager.NewConnectionManager(cfg, metrics.NewMetrics())
```

## Features in Detail

### Circuit Breaker

The circuit breaker prevents cascading failures by stopping requests to a failing service:

```go
// Circuit breaker is automatically enabled when EnableCircuitBreaker is true
// You can customize the circuit breaker config:
cbConfig := interceptors.DefaultCircuitBreakerConfig()
cbConfig.FailureThreshold = 10
cbConfig.Timeout = 60 * time.Second
```

### Retry Logic

Automatic retry with exponential backoff:

```go
retryConfig := interceptors.DefaultRetryConfig()
retryConfig.MaxAttempts = 5
retryConfig.InitialBackoff = 200 * time.Millisecond
```

### Metrics

Prometheus metrics are automatically collected when enabled:

- `grpc_client_requests_total`: Total number of gRPC requests
- `grpc_client_request_duration_seconds`: Request duration histogram
- `grpc_client_connections_active`: Number of active connections
- `grpc_client_connection_state`: Connection state gauge
- `grpc_client_retries_total`: Total retry attempts
- `grpc_client_circuit_breaker_state`: Circuit breaker state

### Health Checks

Check the health of all connections:

```go
health := cm.HealthCheck(ctx)
for service, status := range health {
    fmt.Printf("Service: %s, State: %s, Healthy: %v\n", 
        service, status.State, status.Healthy)
}
```

## Examples

See the `examples/` directory for more detailed examples:

- `examples/basic/` - Basic usage example
- `examples/tls/` - TLS/SSL connection example

## Testing

Run tests with:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test ./... -cover
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Learning Resources & References

This project was developed using the following resources and documentation.

### Core Technologies
- **[gRPC-Go Documentation](https://pkg.go.dev/google.golang.org/grpc)** - Official gRPC Go library documentation
- **[gRPC Best Practices](https://grpc.io/docs/guides/performance/)** - gRPC best practices and patterns

### Design Patterns
- **[Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)** - Martin Fowler's article on Circuit Breaker pattern
- **[Retry Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/retry)** - Retry pattern documentation
- **[Connection Pooling](https://en.wikipedia.org/wiki/Connection_pool)** - Connection pooling concepts

### Libraries & Tools
- **[zap Logger](https://github.com/uber-go/zap)** - Fast, structured, leveled logging in Go
- **[Prometheus Client](https://github.com/prometheus/client_golang)** - Prometheus metrics library for Go
- **[gRPC Interceptors](https://pkg.go.dev/google.golang.org/grpc#UnaryClientInterceptor)** - gRPC interceptor pattern

### Go Best Practices
- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go best practices guide
- **[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)** - Go code review guidelines
- **[Go Testing](https://go.dev/doc/tutorial/add-a-test)** - Go testing documentation

### Articles & Tutorials
- [Building Resilient Microservices with gRPC](https://grpc.io/blog/grpc-load-balancing/)
- [gRPC Connection Management](https://github.com/grpc/grpc-go/blob/master/Documentation/connection-backoff.md)
- [Circuit Breaker Implementation in Go](https://www.alexedwards.net/blog/how-to-make-an-http-request-in-go)

## Acknowledgments

- Built with [gRPC-Go](https://github.com/grpc/grpc-go)
- Uses [zap](https://github.com/uber-go/zap) for logging
- Uses [Prometheus](https://prometheus.io/) for metrics
