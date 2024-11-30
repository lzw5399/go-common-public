package mem

import (
	"context"
	"time"

	"github.com/bluele/gcache"

	ferrors "github.com/lzw5399/go-common-public/library/errors"
	"github.com/lzw5399/go-common-public/library/log"
)

type Cache struct {
	c gcache.Cache
}

func NewCache() *Cache {
	return &Cache{
		c: gcache.New(100000).LRU().Build(),
	}
}

func (m *Cache) Set(k string, v interface{}, d time.Duration) error {
	return m.c.SetWithExpire(k, v, d)
}

func (m *Cache) Get(k string) (interface{}, bool) {
	v, err := m.c.Get(k)
	if err == nil {
		return v, true
	}

	return v, false
}

func (m *Cache) Remove(k string) bool {
	return m.c.Remove(k)
}

// GetOrSet 从缓存中获取数据，如果不存在则从fetcher中获取数据并设置到缓存中
func GetOrSet(ctx context.Context, cache *Cache, cacheKey string, expiration time.Duration, fetcher func(ctx context.Context) (interface{}, *ferrors.SvrRspInfo)) (interface{}, *ferrors.SvrRspInfo) {
	val, ok := cache.Get(cacheKey)
	if ok {
		return val, ferrors.Ok()
	}

	val, rspInfo := fetcher(ctx)
	if !rspInfo.Valid() {
		return nil, rspInfo
	}

	err := cache.Set(cacheKey, val, expiration)
	if err != nil {
		log.Errorc(ctx, "mem GetOrSet Set failed: %s", err)
	}
	return val, ferrors.Ok()
}

// GetOrSetCondition 如果condition为true, 则优先从缓存中获取数据，如果不存在则从fetcher中获取数据并设置到缓存中
func GetOrSetCondition(ctx context.Context, cache *Cache, cacheKey string, expiration time.Duration, condition func() bool, fetcher func(ctx context.Context) (interface{}, *ferrors.SvrRspInfo)) (interface{}, *ferrors.SvrRspInfo) {
	// 判断是否使用缓存
	useCache := condition()
	if !useCache {
		val, rspInfo := fetcher(ctx)
		return val, rspInfo
	}

	return GetOrSet(ctx, cache, cacheKey, expiration, fetcher)
}
