package rate_limit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

//令牌桶算法要点：
//• 有一个人按一定的速率发令牌
//• 令牌会被放到一个桶里
//• 每一个请求从桶里面拿一个令牌
//• 拿到令牌的请求就会被处理
//• 没有拿到令牌的请求就会：
//• 直接被拒绝
//• 阻塞直到拿到令牌或者超时

type TokenBucketLimiter struct {
	tokens chan struct{}
	close  chan struct{}
}

func NewTokenBucketLimiter(cap int, interval time.Duration) *TokenBucketLimiter {
	ch := make(chan struct{}, cap)
	closeCh := make(chan struct{})

	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				select {
				case ch <- struct{}{}:
				default:
					//没人取令牌
				}
			case <-closeCh:
				return
			}
		}
	}()

	return &TokenBucketLimiter{
		tokens: ch,
		close:  closeCh,
	}
}

func (t *TokenBucketLimiter) Close() error {
	close(t.close)
	return nil
}

func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		//要在这里拿到令牌
		select {
		case <-t.close:
			//return nil, errors.New("tokenBucketLimiter is closed")
			return handler(ctx, req)
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.tokens:
			return handler(ctx, req)
		default:
			err := errors.New("限流")
			return nil, err
		}

		return nil, nil
	}
}

//func (t *TokenBucketLimiter) BuildClientInterceptor() grpc.UnaryClientInterceptor {
//	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
//
//	}
//}
