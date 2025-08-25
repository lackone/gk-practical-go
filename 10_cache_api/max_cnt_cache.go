package _0_cache_api

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// 控制键值对数量
type MaxCntCache struct {
	*LocalCache
	cnt    int32
	maxCnt int32
}

func NewMaxCntCache(l *LocalCache, max int32) *MaxCntCache {
	res := &MaxCntCache{
		LocalCache: l,
		maxCnt:     max,
	}

	origin := l.onEvicted

	res.onEvicted = func(k string, v interface{}) {
		atomic.AddInt32(&res.cnt, 1)
		if origin != nil {
			origin(k, v)
		}
	}

	return res
}

func (m *MaxCntCache) Set(ctx context.Context, k string, v any, duration time.Duration) error {
	//如果k已经存在，这里的计数就会不准
	//cnt := atomic.AddInt32(&m.cnt, 1)
	//if cnt > m.maxCnt {
	//	return errors.New("max cnt")
	//}
	//return m.LocalCache.Set(ctx, k, v, duration)

	//可能m.cnt++执行多次，
	//m.mutex.Lock()
	//_, ok := m.data[k]
	//if !ok {
	//	m.cnt++
	//}
	//if m.cnt > m.maxCnt {
	//	m.mutex.Unlock()
	//	return errors.New("max cnt")
	//}
	//m.mutex.Unlock()
	//
	//return m.LocalCache.Set(ctx, k, v, duration)

	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, ok := m.data[k]
	if !ok {
		if m.cnt+1 > m.maxCnt {
			return errors.New("max cnt")
		}
		m.cnt++
	}
	//注意，这里调用的无锁版本set
	return m.LocalCache.set(ctx, k, v, duration)
}
