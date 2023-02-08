package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	goredis "github.com/go-redis/redis/v8"

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
	list, err := r.rdb.Get(context.Background(), r.generateKey(id)).Result()
	results := make([]string, 0)
	if err != nil {
		log.Printf("[ERROR] Failed to set result: %s\n", err)
	}

	json.Unmarshal([]byte(list), &results)

	return results
}

func (r *redisCache) generateKey(id string) string {
	return fmt.Sprintf("%s:%s", r.prefix, id)
}

func NewRedisCache(prefix string, rdb *goredis.Client, ttl time.Duration) Cache {
	return &redisCache{rdb: rdb, prefix: prefix, ttl: ttl}
}
