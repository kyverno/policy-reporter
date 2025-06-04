package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
	"openreports.io/apis/openreports.io/v1alpha1"
)

type inMemoryCache struct {
	caches       *gocache.Cache
	keepDuration time.Duration
	keepReport   time.Duration
}

func (c *inMemoryCache) AddReport(report v1alpha1.ReportInterface) {
	cache, ok := c.getCache(report.GetID())

	if !ok {
		cache = gocache.New(gocache.NoExpiration, 5*time.Minute)
		c.caches.Set(report.GetID(), cache, gocache.NoExpiration)
	}

	next := make(map[string]bool)
	for _, result := range report.GetResults() {
		cache.Set(result.GetID(), nil, gocache.NoExpiration)
		next[result.GetID()] = true
	}

	for id, item := range cache.Items() {
		if !next[id] && item.Expiration == 0 {
			cache.Set(id, nil, c.keepDuration)
		}
	}

	c.caches.Set(report.GetID(), cache, gocache.NoExpiration)
}

func (c *inMemoryCache) RemoveReport(id string) {
	cache, ok := c.getCache(id)
	if !ok {
		return
	}

	c.caches.Set(id, cache, c.keepReport)
}

func (c *inMemoryCache) getCache(id string) (*gocache.Cache, bool) {
	cache, ok := c.caches.Get(id)
	if !ok {
		return nil, false
	}

	return cache.(*gocache.Cache), ok
}

func (c *inMemoryCache) GetResults(id string) []string {
	list := make([]string, 0)

	cache, ok := c.getCache(id)
	if !ok {
		return list
	}

	for id := range cache.Items() {
		list = append(list, id)
	}

	return list
}

func (c *inMemoryCache) Clear() {
	for _, cache := range c.caches.Items() {
		cache.Object.(*gocache.Cache).Flush()
	}

	c.caches.Flush()
}

func (c *inMemoryCache) Shared() bool {
	return false
}

func NewInMermoryCache(keepDuration, keepReport time.Duration) Cache {
	cache := gocache.New(gocache.NoExpiration, 5*time.Minute)
	cache.OnEvicted(func(s string, i interface{}) {
		if c, ok := i.(*gocache.Cache); ok {
			c.Flush()
		}
	})

	return &inMemoryCache{
		caches:       gocache.New(gocache.NoExpiration, 5*time.Minute),
		keepDuration: keepDuration,
		keepReport:   keepReport,
	}
}
