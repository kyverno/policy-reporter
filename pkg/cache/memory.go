package cache

import (
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
)

type inMemoryCache struct {
	mx     *sync.RWMutex
	caches map[string]*gocache.Cache
}

func (c *inMemoryCache) AddReport(report v1alpha2.ReportInterface) {
	cache, ok := c.getCache(report.GetID())

	if !ok {
		cache = gocache.New(gocache.NoExpiration, 5*time.Minute)
		c.addCache(report.GetID(), cache)
	}

	current := c.GetResults(report.GetID())
	next := make([]string, 0, len(report.GetResults()))
	for _, result := range report.GetResults() {
		cache.Set(result.GetID(), nil, gocache.NoExpiration)
		next = append(next, result.GetID())
	}

	for _, id := range current {
		if !helper.Contains(id, next) {
			cache.Set(id, nil, 6*time.Hour)
		}
	}
}

func (c *inMemoryCache) RemoveReport(id string) {
	cache, ok := c.getCache(id)

	if !ok {
		return
	}

	cache.Flush()
	c.mx.Lock()
	delete(c.caches, id)
	c.mx.Unlock()
}

func (c *inMemoryCache) getCache(id string) (*gocache.Cache, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	cache, ok := c.caches[id]

	return cache, ok
}

func (c *inMemoryCache) addCache(id string, cache *gocache.Cache) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.caches[id] = cache
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
	for id, cache := range c.caches {
		cache.Flush()

		c.mx.Lock()
		delete(c.caches, id)
		c.mx.Unlock()
	}
}

func (c *inMemoryCache) Shared() bool {
	return false
}

func NewInMermoryCache() Cache {
	return &inMemoryCache{
		caches: make(map[string]*gocache.Cache),
		mx:     new(sync.RWMutex),
	}
}
