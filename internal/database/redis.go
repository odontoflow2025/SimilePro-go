package database

import (
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis(addr string) {
	if addr == "" {
		addr = "localhost:6379"
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	log.Printf("Redis initialized with address: %s", addr)
}
