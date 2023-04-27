package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type redisCache struct {
	rdb    *goredis.Client
	prefix string
	ttl    time.Duration
}

func (r *redisCache) AddReport(report v1alpha2.ReportInterface) {
	list := reportResultsIds(report)

	value, _ := json.Marshal(list)

	r.rdb.Set(context.Background(), r.generateKey(report.GetID()), string(value), 0)
}

func (r *redisCache) RemoveReport(id string) {
	r.rdb.Del(context.Background(), r.generateKey(id))
}

func (r *redisCache) GetResults(id string) []string {
	list, _ := r.rdb.Get(context.Background(), r.generateKey(id)).Result()
	results := make([]string, 0)

	json.Unmarshal([]byte(list), &results)

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

func (r *redisCache) generateKey(id string) string {
	return fmt.Sprintf("%s:%s", r.prefix, id)
}

func NewRedisCache(prefix string, rdb *goredis.Client, ttl time.Duration) Cache {
	return &redisCache{rdb: rdb, prefix: prefix, ttl: ttl}
}
