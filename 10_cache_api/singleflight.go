package _0_cache_api

import (
	"context"
	"golang.org/x/sync/singleflight"
	"log"
	"time"
)

type SingleflightCache struct {
	ReadThroughCache
}

func NewSingleflightCache(cache CacheV2,
	loadFunc func(ctx context.Context, key string) (any, error),
	duration time.Duration,
) *SingleflightCache {
	g := &singleflight.Group{}
	return &SingleflightCache{
		ReadThroughCache: ReadThroughCache{
			CacheV2: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				v, err, _ := g.Do(key, func() (any, error) {
					return loadFunc(ctx, key)
				})
				return v, err
			},
			Expiration: duration,
		},
	}
}

type SingleflightCacheV2 struct {
	ReadThroughCache
	g *singleflight.Group
}

func (s *SingleflightCacheV2) Get(ctx context.Context, key string, duration time.Duration) (any, error) {
	val, err := s.CacheV2.Get(ctx, key)
	if err != nil {
		val, err, _ = s.g.Do(key, func() (any, error) {
			v, er := s.LoadFunc(ctx, key)
			if er == nil {
				if e := s.Set(ctx, key, v, duration); e != nil {
					log.Fatalln(e.Error())
				}
			}
			return v, err
		})
	}
	return val, err
}
