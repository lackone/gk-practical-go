package _0_cache_api

import (
	"context"
	"fmt"
	"golang.org/x/sync/singleflight"
	"log"
	"time"
)

type ReadThroughCache struct {
	CacheV2
	LoadFunc   func(ctx context.Context, key string) (value any, err error)
	Expiration time.Duration
	g          *singleflight.Group
}

func (r *ReadThroughCache) Get(ctx context.Context, k string) (v any, err error) {
	val, err := r.CacheV2.Get(ctx, k)
	if err != nil {
		val, err = r.LoadFunc(ctx, k)
		if err == nil {
			er := r.CacheV2.Set(ctx, k, val, r.Expiration)
			if er != nil {
				return val, fmt.Errorf("%s", er.Error())
			}
		}
	}
	return val, err
}

// 全异步
func (r *ReadThroughCache) GetV1(ctx context.Context, k string) (v any, err error) {
	val, err := r.CacheV2.Get(ctx, k)
	if err != nil {
		go func() {
			val, err = r.LoadFunc(ctx, k)
			if err == nil {
				er := r.CacheV2.Set(ctx, k, val, r.Expiration)
				if er != nil {
					//return val, fmt.Errorf("%s", er.Error())
					log.Fatalln(er.Error())
				}
			}
		}()
	}
	return val, err
}

// 半异步
func (r *ReadThroughCache) GetV2(ctx context.Context, k string) (v any, err error) {
	val, err := r.CacheV2.Get(ctx, k)
	if err != nil {
		val, err = r.LoadFunc(ctx, k)
		if err == nil {
			go func() {
				er := r.CacheV2.Set(ctx, k, val, r.Expiration)
				if er != nil {
					//return val, fmt.Errorf("%s", er.Error())
					log.Fatalln(er.Error())
				}
			}()
		}
	}
	return val, err
}

func (r *ReadThroughCache) GetV3(ctx context.Context, k string) (v any, err error) {
	val, err := r.CacheV2.Get(ctx, k)
	if err != nil {
		val, err, _ = r.g.Do(k, func() (interface{}, error) {
			val, err = r.LoadFunc(ctx, k)
			if err == nil {
				er := r.CacheV2.Set(ctx, k, val, r.Expiration)
				if er != nil {
					return val, fmt.Errorf("%s", er.Error())
				}
			}
			return val, err
		})
	}
	return val, err
}
