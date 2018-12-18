package circuitbreaker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	errMsgFake         = "fake error"
	errMsgCircuitBreak = "the request performs circuit break"
)

type mockPassedBreaker struct{}

func (*mockPassedBreaker) CallContext(ctx context.Context, fun func() error) error {
	return fun()
}

func passedUnaryInvoker(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	return nil
}

func failedUnaryInvoker(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	return status.Error(codes.Internal, errMsgFake)
}

type mockCircuitBreaker struct{}

func (*mockCircuitBreaker) CallContext(ctx context.Context, fun func() error) error {
	if err := fun(); err != nil {
		return errors.New(errMsgCircuitBreak)
	}
	return nil
}

func TestUnaryClientInterceptor_NotCircuitBreak(t *testing.T) {
	interceptor := UnaryClientInterceptor(&mockPassedBreaker{})
	err := interceptor(nil, "", nil, nil, nil, passedUnaryInvoker)
	assert.NoError(t, err)

	err2 := interceptor(nil, "", nil, nil, nil, failedUnaryInvoker)
	assert.EqualError(t, err2, "rpc error: code = Internal desc = fake error")
}

func TestUnaryClientInterceptor_CircuitBreak(t *testing.T) {
	interceptor := UnaryClientInterceptor(&mockCircuitBreaker{})
	err := interceptor(nil, "", nil, nil, nil, passedUnaryInvoker)
	assert.NoError(t, err)

	err2 := interceptor(nil, "", nil, nil, nil, failedUnaryInvoker)
	assert.EqualError(t, err2, errMsgCircuitBreak)
}

func passedStreamer(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func failedStreamer(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, status.Error(codes.Internal, errMsgFake)
}

func TestStreamClientInterceptor_NotCircuitBreak(t *testing.T) {
	interceptor := StreamClientInterceptor(&mockPassedBreaker{})
	_, err := interceptor(nil, nil, nil, "", passedStreamer)
	assert.NoError(t, err)

	_, err2 := interceptor(nil, nil, nil, "", failedStreamer)
	assert.EqualError(t, err2, "rpc error: code = Internal desc = fake error")
}

func TestStreamClientInterceptor_CircuitBreak(t *testing.T) {
	interceptor := StreamClientInterceptor(&mockCircuitBreaker{})
	_, err := interceptor(nil, nil, nil, "", passedStreamer)
	assert.NoError(t, err)

	_, err2 := interceptor(nil, nil, nil, "", failedStreamer)
	assert.EqualError(t, err2, errMsgCircuitBreak)
}
