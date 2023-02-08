package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type redisCache struct {
	rdb    *goredis.Client
	prefix string
	ttl    time.Duration
}

func (r *redisCache) Add(id string) {
	err := r.rdb.Set(context.Background(), r.generateKey(id), true, r.ttl).Err()
	if err != nil {
		log.Printf("[ERROR] Failed to set result: %s\n", err)
	}
}

func (r *redisCache) Has(id string) bool {
	_, err := r.rdb.Get(context.Background(), r.generateKey(id)).Result()
	if err == goredis.Nil {
		return false
	} else if err != nil {
		log.Printf("[ERROR] Failed to get result: %s\n", err)
		return false
	}

	return true
}

func (r *redisCache) generateKey(id string) string {
	return fmt.Sprintf("%s:%s", r.prefix, id)
}

func NewRedisCache(prefix string, rdb *goredis.Client, ttl time.Duration) Cache {
	return &redisCache{rdb: rdb, prefix: prefix, ttl: ttl}
}
