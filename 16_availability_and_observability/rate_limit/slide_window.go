package rate_limit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	queue    *list.List
	interval int64
	rate     int
	mutex    sync.Mutex
}

func (s *SlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		//我要知道当前的这个窗口，处理了几个请求
		now := time.Now().UnixNano()

		//当前窗口最开始的时间
		start := now - s.interval

		s.mutex.Lock()
		//快路径
		len := s.queue.Len()
		if len < s.rate {
			resp, err := handler(ctx, req)
			s.queue.PushBack(now) //保存了请求的时间戳
			s.mutex.Unlock()
			return resp, err
		}

		//慢路径，删除元素
		tmp := s.queue.Front()
		//把不在窗口的元素删掉，也就是小于start的元素
		for tmp != nil && tmp.Value.(int64) < start {
			s.queue.Remove(tmp)
			tmp = s.queue.Front()
		}

		len = s.queue.Len()

		s.mutex.Unlock()

		if len >= s.rate {
			return nil, errors.New("too many requests")
		}

		resp, err := handler(ctx, req)
		s.queue.PushBack(now) //保存了请求的时间戳
		return resp, err
	}
}
