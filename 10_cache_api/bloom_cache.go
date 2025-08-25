package _0_cache_api

import (
	"context"
	"errors"
)

type BlommFilterCache struct {
	ReadThroughCache
}

func NewBlommFilterCache(cache CacheV2,
	bf BlommFilter,
	loadFunc func(ctx context.Context, key string) (any, error),
) *BlommFilterCache {
	return &BlommFilterCache{
		ReadThroughCache: ReadThroughCache{
			CacheV2: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				if !bf.HasKey(ctx, key) {
					return nil, errors.New("key not found")
				}
				return loadFunc(ctx, key)
			},
		},
	}
}

type BlommFilter interface {
	HasKey(ctx context.Context, key string) bool
}
