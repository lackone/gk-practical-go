package rate_limit

import (
	"context"
	_ "embed"
	"errors"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"time"
)

//go:embed lua/slide_window.lua
var luaSlideWindow string

type RedisSlideWindowLimiter struct {
	client   redis.Cmdable
	interval time.Duration //窗口大小
	rate     int64         //在这个窗口内，允许通过的最大请求数量
}

func NewRedisSlideWindowLimiter(client redis.Cmdable, interval time.Duration, rate int64) *RedisSlideWindowLimiter {
	return &RedisSlideWindowLimiter{
		client:   client,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisSlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 预期 lua 脚本返回一个 bool 值，要不要限流
		limit, err := r.limit(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		if limit {
			return nil, errors.New("too many requests")
		}
		return handler(ctx, req)
	}
}

func (r *RedisSlideWindowLimiter) limit(ctx context.Context, key string) (bool, error) {
	b, err := r.client.Eval(ctx, luaSlideWindow, []string{key}, r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
	return b, err
}
