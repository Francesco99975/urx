package cache

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

var (
	client *redis.Client
	once   sync.Once
)

func InitCache(addr string) {
	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr:         addr, // "valkey:6379"
			DB:           0,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
		})
	})
	// test connection
	_, err := client.Ping(Ctx).Result()
	if err != nil {
		panic(err)
	}
}

func Client() *redis.Client {
	return client
}
