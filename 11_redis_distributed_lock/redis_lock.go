package _1_redis_distributed_lock

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"time"
)

var (
	ErrFailedToPreemptLock = errors.New("failed to preempt lock") //抢锁失败
	ErrLockNotExist        = errors.New("lock not exist")         //锁不存在
	ErrLockNotHold         = errors.New("lock not hold")          //你没有持有锁

	//go:embed unlock.lua
	luaUnlock string

	//go:embed refresh.lua
	luaRefresh string

	//go:embed lock.lua
	luaLock string
)

// Client 就是对 redis.Cmdable 二次封装
type Client struct {
	client redis.Cmdable
	g      singleflight.Group
}

func NewClient(client redis.Cmdable) *Client {
	return &Client{
		client: client,
	}
}

type Lock struct {
	client   redis.Cmdable
	key      string
	val      string
	expire   time.Duration
	unlockCh chan struct{}
}

type RetryStrategy interface {
	//第一个返回值，重试的间隔
	//第二个返回值，要不要重试
	Next() (time.Duration, bool)
}

func (c *Client) SingleflightLock(ctx context.Context,
	key string,
	duration time.Duration,
	timeout time.Duration,
	retry RetryStrategy,
) (*Lock, error) {
	for {
		flag := false

		resCh := c.g.DoChan(key, func() (interface{}, error) {
			flag = true
			return c.Lock(ctx, key, duration, timeout, retry)
		})

		select {
		case res := <-resCh:
			if flag { //拿到了锁
				c.g.Forget(key)
				if res.Err != nil {
					return nil, res.Err
				}
				return res.Val.(*Lock), nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) Lock(ctx context.Context,
	key string,
	duration time.Duration,
	timeout time.Duration,
	retry RetryStrategy,
) (*Lock, error) {

	var timer *time.Timer
	val := uuid.New().String()

	for {
		lctx, lcancel := context.WithTimeout(ctx, timeout)

		res, err := c.client.Eval(lctx, luaLock, []string{key}, val, duration.Seconds()).Result()

		lcancel()

		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		if res == "OK" {
			// 加锁成功
			return &Lock{
				client: c.client,
				key:    key,
				val:    val,
				expire: duration,
			}, nil
		}

		interval, ok := retry.Next()
		if !ok {
			return nil, fmt.Errorf("超出重试限制 %w", ErrFailedToPreemptLock)
		}
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}

		select {
		case <-timer.C:
		case <-ctx.Done():
			//控制的是整体的超时
			return nil, ctx.Err()
		}
	}
}

func (c *Client) TryLock(ctx context.Context, key string, duration time.Duration) (*Lock, error) {
	//生成一个唯一的值
	val := uuid.New().String()
	// SetNX 一定要有过期时间，不然实例崩溃，没人释放锁了
	result, err := c.client.SetNX(ctx, key, val, duration).Result()
	if err != nil {
		return nil, err
	}
	if !result {
		//代表别人抢到了锁
		return nil, ErrFailedToPreemptLock
	}
	return &Lock{
		client: c.client,
		key:    key,
		val:    val,
		expire: duration,
	}, nil
}

// 在解锁的时候，要判断一下是不是自已的锁
func (l *Lock) Unlock(ctx context.Context) error {
	//把键值对删掉
	result, err := l.client.Del(ctx, l.key).Result()
	if err != nil {
		return err
	}
	if result != 1 {
		//代表你加的锁，过期了
		return errors.New("解锁失败，锁不存在")
	}
	return nil
}

func (l *Lock) UnlockV2(ctx context.Context) error {
	result, err := l.client.Get(ctx, l.key).Result()
	if err != nil {
		return err
	}
	if result != l.val {
		return errors.New("锁不是自已的锁")
	}
	//把键值对删掉
	cnt, err := l.client.Del(ctx, l.key).Result()
	if err != nil {
		return err
	}
	if cnt != 1 {
		//代表你加的锁，过期了
		return errors.New("解锁失败，锁不存在")
	}
	return nil
}

// Get 和 Del 两个操作必须原子操作
func (l *Lock) UnlockV3(ctx context.Context) error {
	i, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.val).Int64()
	defer func() {
		//close(l.unlockCh)
		l.unlockCh <- struct{}{}
	}()
	if err == redis.Nil {
		return ErrLockNotHold
	}
	if err != nil {
		return err
	}
	if i != 1 {
		return ErrLockNotHold
	}
	return nil
}

func (l *Lock) Refresh(ctx context.Context) error {
	i, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.val, l.expire.Seconds()).Int64()
	if err != nil {
		return err
	}
	if i != 1 {
		return ErrLockNotHold
	}
	return nil
}

func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	timeoutCh := make(chan struct{})

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				timeoutCh <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-timeoutCh:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				timeoutCh <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-l.unlockCh:
			return nil
		}
	}
}
