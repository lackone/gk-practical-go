package _0_cache_api

import (
	"context"
	"log"
	"time"
)

type WriteThroughCache struct {
	CacheV2
	StoreFunc func(ctx context.Context, key string, value any) error
}

func (w *WriteThroughCache) Set(ctx context.Context, key string, value any, duration time.Duration) error {
	err := w.StoreFunc(ctx, key, value)
	if err != nil {
		return err
	}
	return w.CacheV2.Set(ctx, key, value, duration)
}

// 半异步
func (w *WriteThroughCache) SetV2(ctx context.Context, key string, value any, duration time.Duration) error {
	err := w.StoreFunc(ctx, key, value)
	go func() {
		err := w.CacheV2.Set(ctx, key, value, duration)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}()
	return err
}

// 全异步
func (w *WriteThroughCache) SetV3(ctx context.Context, key string, value any, duration time.Duration) error {
	go func() {
		err := w.StoreFunc(ctx, key, value)
		if err != nil {
			log.Fatalln(err.Error())
		}
		err = w.CacheV2.Set(ctx, key, value, duration)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}()
	return nil
}
