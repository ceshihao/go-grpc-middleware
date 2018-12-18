package circuitbreaker_test

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/circuitbreaker"
	"google.golang.org/grpc"
)

const errMsgCircuitBreak = "the request performs circuit break"

// alwaysCircuitBreaker is an example breaker which impletemts Breaker interface.
// It circuit breakt all requests because CallContext function always returns errMsgCircuitBreak.
type alwaysCircuitBreaker struct{}

func (*alwaysCircuitBreaker) CallContext(ctx context.Context, fun func() error) error {
	return errors.New(errMsgCircuitBreak)
}

// Simple example of grpc Dial code.
func Example() {
	// Create unary/stream circuitBreakers.
	// You can implement your own breaker for the interface.
	breaker := &alwaysCircuitBreaker{}
	grpc.Dial("myservice.example.com",
		grpc.WithStreamInterceptor(circuitbreaker.StreamClientInterceptor(breaker)),
		grpc.WithUnaryInterceptor(circuitbreaker.UnaryClientInterceptor(breaker)),
	)
}
