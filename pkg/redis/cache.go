package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type RedisCache struct {
	rdb    *goredis.Client
	prefix string
	ttl    time.Duration
}

func (r *RedisCache) Add(id string) {
	err := r.rdb.Set(context.Background(), r.generateKey(id), true, r.ttl).Err()
	if err != nil {
		log.Printf("[ERROR] Failed to set result: %s\n", err)
	}
}

func (r *RedisCache) Has(id string) bool {
	_, err := r.rdb.Get(context.Background(), r.generateKey(id)).Result()
	if err == goredis.Nil {
		return false
	} else if err != nil {
		log.Printf("[ERROR] Failed to get result: %s\n", err)
		return false
	}

	return true
}

func (r *RedisCache) generateKey(id string) string {
	return fmt.Sprintf("%s:%s", r.prefix, id)
}

func New(prefix string, rdb *goredis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{rdb: rdb, prefix: prefix, ttl: ttl}
}
