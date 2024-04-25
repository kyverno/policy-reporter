package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type rdb interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
	Keys(ctx context.Context, pattern string) *goredis.StringSliceCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *goredis.BoolCmd
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
}

type redisCache struct {
	rdb    rdb
	prefix string
	ttl    time.Duration
}

func (r *redisCache) AddReport(report v1alpha2.ReportInterface) {
	next := make(map[string]bool)

	for _, result := range report.GetResults() {
		r.rdb.Set(context.Background(), r.generateKey(report.GetID(), result.GetID()), nil, 0)
		next[result.GetID()] = true
	}

	for _, id := range r.GetResults(report.GetID()) {
		if !next[id] {
			r.rdb.Set(context.Background(), r.generateKey(report.GetID(), id), nil, r.ttl)
		}
	}
}

func (r *redisCache) RemoveReport(id string) {
	keys, err := r.rdb.Keys(context.Background(), r.generateKeyPattern(id)).Result()
	if err != nil {
		zap.L().Error("failed to load report keys", zap.Error(err))
		return
	}

	for _, key := range keys {
		r.rdb.Expire(context.Background(), key, 10*time.Minute)
	}
}

func (r *redisCache) GetResults(id string) []string {
	results := make([]string, 0)
	pattern := r.generateKeyPattern(id)

	keys, err := r.rdb.Keys(context.Background(), pattern).Result()
	if err != nil {
		zap.L().Error("failed to load report keys", zap.Error(err))
		return results
	}

	prefix := strings.TrimSuffix(pattern, "*")
	for _, key := range keys {
		results = append(results, strings.TrimPrefix(key, prefix))
	}

	return results
}

func (r *redisCache) Shared() bool {
	return true
}

func (r *redisCache) Clear() {
	results := r.rdb.Keys(context.Background(), r.prefix+":*")

	keys, err := results.Result()
	if err != nil {
		zap.L().Error("failed to find cache keys in redis", zap.Error(err))
	}

	for _, key := range keys {
		r.rdb.Del(context.Background(), key)
	}
}

func (r *redisCache) generateKey(report, id string) string {
	return fmt.Sprintf("%s:%s:%s", r.prefix, report, id)
}

func (r *redisCache) generateKeyPattern(report string) string {
	return fmt.Sprintf("%s:%s:*", r.prefix, report)
}

func NewRedisCache(prefix string, rdb rdb, ttl time.Duration) Cache {
	return &redisCache{rdb: rdb, prefix: prefix, ttl: ttl}
}
