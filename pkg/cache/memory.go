package cache

import (
	"time"

	gocache "zgo.at/zcache/v2"

	"github.com/kyverno/policy-reporter/pkg/openreports"
)

type inMemoryCache struct {
	caches       *gocache.Cache[string, *gocache.Cache[string, any]]
	keepDuration time.Duration
	keepReport   time.Duration
}

func (c *inMemoryCache) AddReport(report openreports.ReportInterface) {
	cache, ok := c.getCache(report.GetID())

	if !ok {
		cache = gocache.New[string, any](gocache.NoExpiration, 5*time.Minute)
		c.caches.Set(report.GetID(), cache)
	}

	next := make(map[string]bool)
	for _, result := range report.GetResults() {
		cache.Set(result.GetID(), nil)
		next[result.GetID()] = true
	}

	for id, item := range cache.Items() {
		if !next[id] && item.Expiration == 0 {
			cache.SetWithExpire(id, nil, c.keepDuration)
		}
	}

	c.caches.Set(report.GetID(), cache)
}

func (c *inMemoryCache) RemoveReport(id string) {
	cache, ok := c.getCache(id)
	if !ok {
		return
	}

	c.caches.SetWithExpire(id, cache, c.keepReport)
}

func (c *inMemoryCache) getCache(id string) (*gocache.Cache[string, any], bool) {
	return c.caches.Get(id)
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
		cache.Object.Reset()
	}

	c.caches.Reset()
}

func (c *inMemoryCache) Shared() bool {
	return false
}

func NewInMemoryCache(keepDuration, keepReport time.Duration) Cache {
	cache := gocache.New[string, *gocache.Cache[string, any]](gocache.NoExpiration, 5*time.Minute)
	cache.OnEvicted(func(s string, c *gocache.Cache[string, any]) {
		c.Reset()
	})

	return &inMemoryCache{
		caches:       cache,
		keepDuration: keepDuration,
		keepReport:   keepReport,
	}
}
