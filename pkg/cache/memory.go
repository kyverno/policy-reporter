package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type inMemoryCache struct {
	cache *gocache.Cache
}

func (c *inMemoryCache) AddReport(report v1alpha2.ReportInterface) {
	list := make([]string, 0, len(report.GetResults()))
	for _, result := range report.GetResults() {
		list = append(list, result.GetID())
	}

	c.cache.Set(report.GetID(), list, gocache.NoExpiration)
}

func (c *inMemoryCache) RemoveReport(id string) {
	val, ok := c.cache.Get(id)
	if ok {
		// don't remove it directly to prevent sending results from instantly recreated reports
		c.cache.Set(id, val, 5*time.Minute)
	}
}

func (c *inMemoryCache) GetResults(id string) []string {
	list, ok := c.cache.Get(id)
	if !ok {
		return make([]string, 0)
	}

	return list.([]string)
}

func (c *inMemoryCache) AddItem(key string, value interface{}) {
	c.cache.Set(key, value, gocache.NoExpiration)
}

func (c *inMemoryCache) RemoveItem(key string) {
	c.cache.Delete(key)
}

func (c *inMemoryCache) GetItem(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func NewInMermoryCache() ItemCache {
	return &inMemoryCache{
		cache: gocache.New(gocache.NoExpiration, 5*time.Minute),
	}
}
