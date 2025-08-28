package rate_limit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"sync/atomic"
	"time"
)

type FixedWindowLimiter struct {
	timestamp int64 //时间戳，窗口的起始时间
	interval  int64 //窗口大小
	rate      int64 //在这个窗口内，允许通过的最大请求数量
	cnt       int64 //当前请求数
	mutex     sync.Mutex
}

func (f *FixedWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		//f.mutex.Lock()

		//考虑 t.cnt 重置的问题
		cur := time.Now().UnixNano()
		timestamp := atomic.LoadInt64(&f.timestamp)
		cnt := atomic.LoadInt64(&f.cnt)

		if timestamp+f.interval < cur {
			if atomic.CompareAndSwapInt64(&f.timestamp, timestamp, cur) {
				atomic.CompareAndSwapInt64(&f.cnt, cnt, 0)
			}
		}

		cnt = atomic.AddInt64(&f.cnt, 1)
		defer atomic.AddInt64(&f.cnt, -1)

		if cnt >= f.rate {
			//f.mutex.Unlock()
			return nil, errors.New("too many requests")
		}

		//f.mutex.Unlock()

		return handler(ctx, req)
	}
}
