package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type Cache interface {
	Has(id string) bool
	Add(id string)
}

type InMemoryCache struct {
	cache *gocache.Cache
}

func (c *InMemoryCache) Has(id string) bool {
	_, ok := c.cache.Get(id)

	return ok
}

func (c *InMemoryCache) Add(id string) {
	c.cache.SetDefault(id, true)
}

func New(defaultExpiration, cleanupInterval time.Duration) *InMemoryCache {
	return &InMemoryCache{
		cache: gocache.New(defaultExpiration, cleanupInterval),
	}
}
