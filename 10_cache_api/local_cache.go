package _0_cache_api

import (
	"context"
	"errors"
	"sync"
	"time"
)

type LocalCache struct {
	data      map[string]*Item
	mutex     sync.RWMutex
	close     chan struct{}
	onEvicted func(key string, val any)
}

type LocalCacheOption func(*LocalCache)

type Item struct {
	data     any
	deadline time.Time
}

func (i *Item) deadlineBefore(t time.Time) bool {
	return !i.deadline.IsZero() && i.deadline.Before(t)
}

func NewLocalCache(duration time.Duration, opt ...LocalCacheOption) *LocalCache {
	res := &LocalCache{
		data: make(map[string]*Item),
	}

	for _, opt := range opt {
		opt(res)
	}

	go func() {
		ticker := time.NewTicker(duration)
		for {
			select {
			case <-ticker.C:
				i := 0
				res.mutex.Lock()
				for k, v := range res.data {
					if i > 1000 { //防止遍历太多
						break
					}
					if v.deadlineBefore(time.Now()) {
						res.delete(k)
					}
					i++
				}
				res.mutex.Unlock()
			case <-res.close:
				return
			}
		}
	}()

	return res
}

func withOnEvicted(f func(key string, val any)) LocalCacheOption {
	return func(res *LocalCache) {
		res.onEvicted = f
	}
}

func (l *LocalCache) delete(key string) {
	v, ok := l.data[key]
	if !ok {
		return
	}
	delete(l.data, key)
	l.onEvicted(key, v.data)
}

func (l *LocalCache) Close() error {
	select {
	case l.close <- struct{}{}:
	default:
		return errors.New("local cache is already closed")
	}
	return nil
}

// 单独抽出一个无锁的
func (l *LocalCache) set(ctx context.Context, k string, v any, duration time.Duration) error {
	var dl time.Time
	if duration > 0 {
		dl = time.Now().Add(duration)
	}

	l.data[k] = &Item{
		data:     v,
		deadline: dl,
	}
	return nil
}

func (l *LocalCache) Set(ctx context.Context, k string, v any, duration time.Duration) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.set(ctx, k, v, duration)

	//每设置一个key，都创建了一个time.AfterFunc
	if duration > 0 {
		time.AfterFunc(duration, func() {
			l.mutex.Lock()
			defer l.mutex.Unlock()

			val, ok := l.data[k]
			if ok && val.deadlineBefore(time.Now()) {
				l.delete(k)
			}

		})
	}

	return nil
}

func (l *LocalCache) Get(ctx context.Context, k string) (v any, err error) {
	l.mutex.RLock()
	val, ok := l.data[k] //先读锁检查
	defer l.mutex.RUnlock()

	if !ok {
		return nil, errors.New("not found")
	}

	now := time.Now()

	if val.deadlineBefore(now) { //判断是否过期
		l.mutex.Lock()
		val, ok = l.data[k]
		if !ok {
			return nil, errors.New("not found")
		}
		if val.deadlineBefore(now) { //再次判断
			l.delete(k)
			return nil, errors.New("timeout")
		}
		l.mutex.Unlock()
		return nil, errors.New("timeout")
	}

	return val.data, nil
}

func (l *LocalCache) Delete(ctx context.Context, k string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.delete(k)
	return nil
}

// 本地实现，直接加锁就行了
func (l *LocalCache) LoadAndDelete(ctx context.Context, k string) (v any, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	val, ok := l.data[k]
	if !ok {
		return nil, errors.New("not found")
	}
	l.delete(k)
	return val.data, nil
}
