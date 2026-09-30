package cache

import (
	"log"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(url string) *redis.Client {
	options, err := redis.ParseURL(url)
	if err != nil {
		log.Fatalf("parse redis url: %v", err)
	}

	client := redis.NewClient(options)

	return client
}
