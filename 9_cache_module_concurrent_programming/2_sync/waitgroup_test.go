package __sync

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestWaitGroup(t *testing.T) {
	wg := sync.WaitGroup{}
	var ret int32 = 0

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(ix int) {
			defer wg.Done()
			atomic.AddInt32(&ret, 1)
		}(i)
	}

	wg.Wait()
	t.Log(ret)
}
