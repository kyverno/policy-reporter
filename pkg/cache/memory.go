package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type inMemoryCache struct {
	cache *gocache.Cache
}

func (c *inMemoryCache) Has(id string) bool {
	_, ok := c.cache.Get(id)

	return ok
}

func (c *inMemoryCache) Add(id string) {
	c.cache.SetDefault(id, true)
}

func NewInMermoryCache(defaultExpiration, cleanupInterval time.Duration) Cache {
	return &inMemoryCache{
		cache: gocache.New(defaultExpiration, cleanupInterval),
	}
}
