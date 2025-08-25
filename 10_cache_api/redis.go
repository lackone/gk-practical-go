package _0_cache_api

import (
	"context"
	"errors"
	"fmt"
	redis "github.com/redis/go-redis/v9"
	"time"
)

var (
	RedisSetError = errors.New("redis写入失败")
)

type RedisCache struct {
	client redis.Cmdable
}

type RedisConfig struct {
	Addr string
}

//func NewRedisCacheV2(cfg RedisConfig) *RedisCache {
//	return redis.NewClient()
//}

func NewRedisCache(client redis.Cmdable) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

func (r *RedisCache) Set(ctx context.Context, k string, v any, duration time.Duration) error {
	result, err := r.client.Set(ctx, k, v, duration).Result()
	if err != nil {
		return err
	}
	if result != "OK" {
		return fmt.Errorf("%w 返回信息 %s", RedisSetError, result)
	}
	return nil
}

func (r *RedisCache) Get(ctx context.Context, k string) (v any, err error) {
	return r.client.Get(ctx, k).Result()
}

func (r *RedisCache) Delete(ctx context.Context, k string) error {
	_, err := r.client.Del(ctx, k).Result()
	return err
}

func (r *RedisCache) LoadAndDelete(ctx context.Context, k string) (v any, err error) {
	return r.client.GetDel(ctx, k).Result()
}
