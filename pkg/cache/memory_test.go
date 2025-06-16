package cache_test

import (
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

func TestInMemory(t *testing.T) {
	t.Run("add report", func(t *testing.T) {
		id := fixtures.DefaultPolicyReport.GetID()

		c := cache.NewInMermoryCache(time.Millisecond, time.Millisecond)

		c.AddReport(&openreports.ORReportAdapter{Report: fixtures.DefaultPolicyReport})

		results := c.GetResults(id)
		if len(results) != len(fixtures.DefaultPolicyReport.Results) {
			t.Error("expected all results were cached")
		}

		c.AddReport(&openreports.ORReportAdapter{Report: fixtures.MinPolicyReport})

		time.Sleep(3 * time.Millisecond)

		changed := c.GetResults(id)
		if len(changed) != len(fixtures.MinPolicyReport.Results) {
			t.Error("expected all old results were removed")
		}
	})
	t.Run("remove report", func(t *testing.T) {
		id := fixtures.DefaultPolicyReport.GetID()

		c := cache.NewInMermoryCache(time.Millisecond, time.Millisecond)

		c.AddReport(&openreports.ORReportAdapter{Report: fixtures.DefaultPolicyReport})

		c.RemoveReport(id)

		time.Sleep(3 * time.Millisecond)

		results := c.GetResults(id)
		if len(results) != 0 {
			t.Error("expected all results were removed")
		}
	})
	t.Run("ceanup report", func(t *testing.T) {
		id := fixtures.DefaultPolicyReport.GetID()

		c := cache.NewInMermoryCache(time.Millisecond, time.Millisecond)

		c.AddReport(&openreports.ORReportAdapter{Report: fixtures.DefaultPolicyReport})

		c.Clear()

		results := c.GetResults(id)
		if len(results) != 0 {
			t.Error("expected all results were cleaned up")
		}
	})
	t.Run("shared cache", func(t *testing.T) {
		c := cache.NewInMermoryCache(time.Millisecond, time.Millisecond)
		if c.Shared() {
			t.Error("expected in memory cache is not shared")
		}
	})
}
