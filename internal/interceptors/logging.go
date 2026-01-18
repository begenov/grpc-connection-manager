package interceptors

import (
	"context"
	"grpc-connection-manager/pkg/logger"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs gRPC unary calls with timing and error information.
func LoggingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()

	err := invoker(ctx, method, req, reply, cc, opts...)

	duration := time.Since(start)

	if err != nil {
		st, _ := status.FromError(err)
		logger.Warnf("gRPC call failed: method=%s, duration=%v, code=%s, error=%v",
			method, duration, st.Code(), err)
	} else {
		logger.Debugf("gRPC call success: method=%s, duration=%v", method, duration)
	}

	return err
}

// LoggingStreamInterceptor logs gRPC stream calls with timing and error information.
func LoggingStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	start := time.Now()

	stream, err := streamer(ctx, desc, cc, method, opts...)

	duration := time.Since(start)

	if err != nil {
		st, _ := status.FromError(err)
		logger.Warnf("gRPC stream failed: method=%s, duration=%v, code=%s, error=%v",
			method, duration, st.Code(), err)
	} else {
		logger.Debugf("gRPC stream success: method=%s, duration=%v", method, duration)
	}

	return stream, err
}
