package _1_redis_distributed_lock

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	LoadAndDelete(ctx context.Context, key string) ([]byte, error)
	OnEvicted(func(string, []byte))
}

type MaxMemoryCache struct {
	Cache Cache
	max   int64
	used  int64
	mutex *sync.Mutex

	keys *list.LinkedList
}

func NewMaxMemoryCache(max int64, cache Cache) *MaxMemoryCache {
	res := &MaxMemoryCache{
		Cache: cache,
		max:   max,
		used:  0,
		mutex: &sync.Mutex{},
		keys:  list.LinkedList[string]{},
	}
	res.Cache.OnEvicted(res.evicted)
	return res
}

func (m *MaxMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	val, err := m.Cache.Get(ctx, key)
	if err == nil {
		//重新把key移到最新位置
		m.deleteKey(key)
		m.keys.Append(key)
	}

	return val, err
}

func (m *MaxMemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Cache.LoadAndDelete(ctx, key)

	for m.used+int64(len(value)) > m.max {
		k, err := m.keys.Get(0)
		if err != nil {
			return err
		}
		m.Cache.Delete(ctx, k)
	}

	err := m.Cache.Set(ctx, key, value, ttl)
	if err == nil {
		m.used = m.used + int64(len(value))
		m.keys.Append(key)
	}

	return nil
}

func (m *MaxMemoryCache) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Cache.Delete(ctx, key)
}

func (m *MaxMemoryCache) LoadAndDelete(ctx context.Context, key string) ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.Cache.LoadAndDelete(ctx, key)
}

func (m *MaxMemoryCache) OnEvicted(fn func(string, []byte)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Cache.OnEvicted(func(key string, data []byte) {
		m.evicted(key, data)
		fn(key, data)
	})
}

func (m *MaxMemoryCache) evicted(key string, val []byte) {
	m.used = m.used - int64(len(val))
	m.deleteKey(key)
}

func (m *MaxMemoryCache) deleteKey(key string) {
	for i := 0; i < m.keys.Len(); i++ {
		el, _ := m.keys.Get(i)
		if el == key {
			_, _ := m.keys.Delete(i)
			return
		}
	}
}
