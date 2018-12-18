package circuitbreaker

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Breaker defines the interface to perform request circuit breaker.
// When meet the circuit break condition, the request is circuit break.
// Otherwise, the request will pass to server side.
type Breaker interface {
	CallContext(context.Context, func() error) error
}

// UnaryClientInterceptor returns a new unary client interceptor that performs request circuit break.
func UnaryClientInterceptor(cb Breaker) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return cb.CallContext(
			ctx,
			func() error {
				if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
					if s, ok := status.FromError(err); ok {
						if s.Code() == codes.Internal {
							return err
						}
					}
				}
				return nil
			},
		)
	}
}

// StreamClientInterceptor returns a new stream client interceptor that performs request circuit break.
func StreamClientInterceptor(cb Breaker) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		var clientStream grpc.ClientStream
		if err := cb.CallContext(
			ctx,
			func() error {
				var err error
				if clientStream, err = streamer(ctx, desc, cc, method, opts...); err != nil {
					if s, ok := status.FromError(err); ok {
						if s.Code() == codes.Internal {
							return err
						}
					}
				}
				return nil
			},
		); err != nil {
			return nil, err
		}
		return clientStream, nil
	}
}
