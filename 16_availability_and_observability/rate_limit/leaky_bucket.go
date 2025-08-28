package rate_limit

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

//漏桶算法要点：
//• 请求过来先排队
//• 每隔一段时间，放过去一个请求
//• 请求排队直到通过，或者超时

type LeakyBucketLimiter struct {
	producer *time.Ticker
}

func NewLeakyBucketLimiter(producer time.Duration) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		producer: time.NewTicker(producer),
	}
}

func (l *LeakyBucketLimiter) Close() error {
	l.producer.Stop()
	return nil
}

func (l *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-l.producer.C:
			return handler(ctx, req)
		}
	}
}
