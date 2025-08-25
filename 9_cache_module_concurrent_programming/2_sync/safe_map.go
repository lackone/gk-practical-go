package __sync

import "sync"

type SafeMap[K comparable, V any] struct {
	data map[K]V
	lock sync.RWMutex
}

func (s *SafeMap[K, V]) Put(key K, val V) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[key] = val
}

func (s *SafeMap[K, V]) Get(key K) (val V, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	val, ok = s.data[key]
	return val, ok
}

func (s *SafeMap[K, V]) LoadOrStore(key K, newVal V) (val V, loaded bool) {
	s.lock.RLock()
	res, ok := s.data[key]
	s.lock.RUnlock()
	if ok {
		return res, true
	}

	// 注意这里会有问题，如果2个goroutine都走到了这里
	// 都会对map进行设置，并返回 false

	s.lock.Lock()
	defer s.lock.Unlock()

	//这里还要进行2次判断
	res, ok = s.data[key]
	if ok {
		return res, true
	}

	s.data[key] = newVal

	return newVal, false
}
