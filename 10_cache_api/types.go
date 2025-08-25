package _0_cache_api

import (
	"context"
	"time"
)

type Cache interface {
	Set(k string, v any, duration time.Duration) error
	Get(k string) (v any, err error)
	Delete(k string) error
}

type CacheV2 interface {
	Set(ctx context.Context, k string, v any, duration time.Duration) error
	Get(ctx context.Context, k string) (v any, err error)
	Delete(ctx context.Context, k string) error
	LoadAndDelete(ctx context.Context, k string) (v any, err error)
}

// 泛型设计的优缺点：
// 优点：用户不需要做类型断言
// 缺点：一个 CacheV3 的实例，只能存储 T 类型的数据，例如 User 或者 Order
type CacheV3[T any] interface {
	Set(ctx context.Context, k string, v T, duration time.Duration) error
	Get(ctx context.Context, k string) (v T, err error)
	Delete(ctx context.Context, k string) error
}
