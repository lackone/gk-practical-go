package __task_pool

import (
	"context"
	"sync"
)

type Task func() error

type TaskPool struct {
	tasks chan Task
	//close *atomic.Bool
	close     chan struct{}
	closeOnce sync.Once
}

// g 是 goroutine 数量
func NewTaskPool(g int, cap int) *TaskPool {
	res := &TaskPool{
		tasks: make(chan Task, cap),
		//close: &atomic.Bool{},
		close: make(chan struct{}),
	}
	for i := 0; i < g; i++ {
		go func() {

			for {
				select {
				case <-res.close:
					return
				case t := <-res.tasks:
					t()
				}
			}

			//for t := range res.tasks {
			//	if res.close.Load() {
			//		return
			//	}
			//	t()
			//}
		}()
	}
	return res
}

func (t *TaskPool) Submit(ctx context.Context, task Task) error {
	select {
	case t.tasks <- task:
	case <-ctx.Done():
		//代表超时
		return ctx.Err()
	}
	return nil
}

func (t *TaskPool) Close() error {
	//t.close.Store(true)

	//这种写法不行，因为只会有一个goroutine收到
	//t.close <- struct{}{}

	close(t.close)

	//t.closeOnce.Do(func() {
	//	close(t.close)
	//})

	return nil
}
