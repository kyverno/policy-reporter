package metrics

import (
	"strconv"

	gocache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/fasthash/fnv1a"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

type CacheItem struct {
	Labels prometheus.Labels
	Value  float64
}

type Cache struct {
	cache          *gocache.Cache
	filter         *report.ResultFilter
	labelGenerator LabelGenerator
}

func (c *Cache) AddReport(polr v1alpha2.ReportInterface) {
	labels := map[string]*CacheItem{}
	for _, res := range polr.GetResults() {
		if !c.filter.Validate(res) {
			continue
		}

		l := c.labelGenerator(polr, res)

		hash := labelHash(l)

		item, ok := labels[hash]
		if !ok {
			labels[hash] = &CacheItem{
				Labels: l,
				Value:  1,
			}
		} else {
			item.Value = item.Value + 1
		}
	}

	list := make([]*CacheItem, 0, len(labels))
	for _, l := range labels {
		list = append(list, l)
	}

	c.cache.Set(polr.GetID(), list, gocache.NoExpiration)
}

func (c *Cache) Remove(id string) {
	c.cache.Delete(id)
}

func (c *Cache) GetReportLabels(id string) []*CacheItem {
	if item, ok := c.cache.Get(id); ok {
		return item.([]*CacheItem)
	}

	return []*CacheItem{{
		Labels: make(prometheus.Labels),
		Value:  0,
	}}
}

func labelHash(labels prometheus.Labels) string {
	h1 := fnv1a.Init64
	for i, v := range labels {
		h1 = fnv1a.AddString64(h1, i+":"+v)
	}

	return strconv.FormatUint(h1, 10)
}

func NewCache(filter *report.ResultFilter, labelGenerator LabelGenerator) *Cache {
	return &Cache{
		cache:          gocache.New(gocache.NoExpiration, gocache.NoExpiration),
		filter:         filter,
		labelGenerator: labelGenerator,
	}
}
