package cache_test

import (
	"context"
	"strings"
	"testing"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
)

type redis struct {
	items map[string]any
}

func (r *redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	if expiration == -1 {
		delete(r.items, key)
		return nil
	}

	r.items[key] = value
	return nil
}

func (r *redis) Keys(ctx context.Context, pattern string) *goredis.StringSliceCmd {
	p := strings.TrimRight(pattern, "*")

	keys := make([]string, 0, len(r.items))
	for k := range r.items {
		if strings.HasPrefix(k, p) {
			keys = append(keys, k)
		}
	}

	s := &goredis.StringSliceCmd{}
	s.SetVal(keys)

	return s
}

func (r *redis) Expire(ctx context.Context, key string, expiration time.Duration) *goredis.BoolCmd {
	delete(r.items, key)

	return nil
}

func (r *redis) Del(ctx context.Context, keys ...string) *goredis.IntCmd {
	for _, k := range keys {
		delete(r.items, k)
	}

	return nil
}

func newRedis() *redis {
	return &redis{items: make(map[string]any)}
}

func TestRedisCache(t *testing.T) {
	t.Run("add report", func(t *testing.T) {
		id := fixtures.DefaultPolicyReport.GetID()

		c := cache.NewRedisCache("cache", newRedis(), -1)

		c.AddReport(fixtures.DefaultPolicyReport)

		results := c.GetResults(id)
		if len(results) != len(fixtures.DefaultPolicyReport.Results) {
			t.Error("expected all results were cached")
		}

		c.AddReport(fixtures.MinPolicyReport)

		changed := c.GetResults(id)
		if len(changed) != len(fixtures.MinPolicyReport.Results) {
			t.Error("expected all old results were removed")
		}
	})
	t.Run("remove report", func(t *testing.T) {
		id := fixtures.DefaultPolicyReport.GetID()

		c := cache.NewRedisCache("cache", newRedis(), -1)

		c.AddReport(fixtures.DefaultPolicyReport)

		c.RemoveReport(id)

		results := c.GetResults(id)
		if len(results) != 0 {
			t.Error("expected all results were removed")
		}
	})
	t.Run("ceanup report", func(t *testing.T) {
		id := fixtures.DefaultPolicyReport.GetID()

		c := cache.NewRedisCache("cache", newRedis(), -1)

		c.AddReport(fixtures.DefaultPolicyReport)

		c.Clear()

		results := c.GetResults(id)
		if len(results) != 0 {
			t.Error("expected all results were cleaned up")
		}
	})
	t.Run("shared cache", func(t *testing.T) {
		c := cache.NewRedisCache("cache", newRedis(), -1)
		if !c.Shared() {
			t.Error("expected redis cache is shared")
		}
	})
}
