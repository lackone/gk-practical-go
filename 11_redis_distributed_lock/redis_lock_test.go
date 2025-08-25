package _1_redis_distributed_lock

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestRedisLock(t *testing.T) {

}

func TestRefresh(t *testing.T) {
	var l *Lock //如果这里加锁成功

	stop := make(chan struct{})
	errCh := make(chan error)
	timeoutCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Second)
		timeoutRetry := 0
		for {
			select {
			case <-ticker.C:
				//刷新的超时时间怎么设置
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				//出了error了怎么办
				err := l.Refresh(ctx)
				cancel()

				//如果超时
				if err == context.DeadlineExceeded {
					timeoutCh <- struct{}{}
					continue
				}

				if err != nil {
					errCh <- err
					return
				}

				//续约成功
				timeoutRetry = 0
			case <-timeoutCh:
				timeoutRetry++
				if timeoutRetry > 10 {
					errCh <- context.DeadlineExceeded
					return
				}
				//刷新的超时时间怎么设置
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				//出了error了怎么办
				err := l.Refresh(ctx)
				cancel()

				//如果超时
				if err == context.DeadlineExceeded {
					timeoutCh <- struct{}{}
					continue
				}

				if err != nil {
					errCh <- err
					return
				}
			case <-stop:
				return
			}
		}
	}()

	//假设这里是你的业务代码
	select {
	case err := <-errCh:
		//续约失败，中断业务
		log.Fatalln(err)
		return
	default:
		// 这是你的步骤1
	}

	//业务退出了，要退出续约的循环
	stop <- struct{}{}
	//l.Unlock()
}
