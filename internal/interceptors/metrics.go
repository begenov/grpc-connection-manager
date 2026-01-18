package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"grpc-connection-manager/internal/metrics"
)

// MetricsInterceptor creates a metrics interceptor for gRPC unary calls.
// It records request counts, durations, and error codes to Prometheus metrics.
func MetricsInterceptor(serviceName string, m *metrics.Metrics) grpc.UnaryClientInterceptor {
	if m == nil {
		return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start)
		code := "OK"

		if err != nil {
			st, _ := status.FromError(err)
			code = st.Code().String()
		}

		m.RecordGRPCRequest(serviceName, method, code, duration)

		return err
	}
}

// MetricsStreamInterceptor creates a metrics interceptor for gRPC stream calls.
// It records request counts, durations, and error codes to Prometheus metrics.
func MetricsStreamInterceptor(serviceName string, m *metrics.Metrics) grpc.StreamClientInterceptor {
	if m == nil {
		return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return streamer(ctx, desc, cc, method, opts...)
		}
	}

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()

		stream, err := streamer(ctx, desc, cc, method, opts...)

		duration := time.Since(start)
		code := "OK"

		if err != nil {
			st, _ := status.FromError(err)
			code = st.Code().String()
		}

		m.RecordGRPCRequest(serviceName, method, code, duration)

		return stream, err
	}
}
