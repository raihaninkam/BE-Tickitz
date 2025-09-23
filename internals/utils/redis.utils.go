package utils

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetFromCache - generic ambil data slice dari Redis
func GetFromCache[T any](ctx context.Context, rdb *redis.Client, key string) ([]T, bool) {
	cmd := rdb.Get(ctx, key)
	if cmd.Err() == nil {
		var result []T
		if b, err := cmd.Bytes(); err == nil {
			if err := json.Unmarshal(b, &result); err == nil {
				if len(result) > 0 {
					return result, true
				}
			}
		}
	} else if cmd.Err() != redis.Nil {
		log.Println("Redis Error.\nCause:", cmd.Err().Error())
	}
	return nil, false
}

// SetToCache - simpan data slice ke Redis dengan TTL
func SetToCache[T any](ctx context.Context, rdb *redis.Client, key string, value []T, ttl time.Duration) {
	if b, err := json.Marshal(value); err == nil {
		if err := rdb.Set(ctx, key, b, ttl).Err(); err != nil {
			log.Println("Redis Error saat set.\nCause:", err.Error())
		}
	} else {
		log.Println("Marshal Error.\nCause:", err.Error())
	}
}

func InvalidateCache(ctx context.Context, rdb *redis.Client, keys ...string) error {
	for _, key := range keys {
		if err := rdb.Del(ctx, key).Err(); err != nil {
			return err
		}
	}
	return nil
}
