package middleware

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/eviltomorrow/open-terminal/lib/zlog"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	hystrix.ConfigureCommand("grpc_client", hystrix.CommandConfig{
		Timeout:                1000 * 10, // 超时时间1000ms
		MaxConcurrentRequests:  100,       // 最大并发数40
		RequestVolumeThreshold: 50,        // 请求数量阙值20，达到这个阙值才可能触发熔断
		ErrorPercentThreshold:  50,        // 错误百分比例阙值 20%
	})
}

func UnaryClientCircuitbreakerInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return hystrix.Do("grpc_client", func() error {
		return invoker(ctx, method, req, reply, cc, opts...)
	}, func(err error) error {
		if err != nil {
			zlog.Error("circuitbreaker was wrong", zap.Error(err), zap.String("target", cc.Target()), zap.String("method", method))
		}
		return nil
	})
}

type wrappedClientStream struct {
	grpc.ClientStream
}

func (w *wrappedClientStream) RecvMsg(m interface{}) error {
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedClientStream) SendMsg(m interface{}) error {
	return w.ClientStream.SendMsg(m)
}

// StreamServerRecoveryInterceptor recover
func StreamClientCircuitbreakerInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}

	return &wrappedClientStream{s}, nil
}
