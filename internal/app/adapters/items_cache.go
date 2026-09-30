package adapters

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"cosy-console/internal/domain/catalog/models"
)

const (
	activeItemsKey = "catalog:active_items"
	activeItemsTTL = 5 * time.Minute
)

// RedisItemsCache implements the catalog ActiveItemsCache port. Redis
// failures degrade to cache misses: Fetch returns found=false, Store logs
// and moves on.
type RedisItemsCache struct {
	client *redis.Client
}

func NewRedisItemsCache(client *redis.Client) *RedisItemsCache {
	return &RedisItemsCache{
		client: client,
	}
}

func (c *RedisItemsCache) Fetch(ctx context.Context) ([]models.Item, bool) {
	data, err := c.client.Get(ctx, activeItemsKey).Bytes()
	if err != nil {
		return nil, false
	}

	items := make([]models.Item, 0)

	if err := json.Unmarshal(data, &items); err != nil {
		return nil, false
	}

	return items, true
}

func (c *RedisItemsCache) Store(ctx context.Context, items []models.Item) {
	data, err := json.Marshal(items)
	if err != nil {
		log.Printf("cache items: marshal: %v", err)

		return
	}

	if err := c.client.Set(ctx, activeItemsKey, data, activeItemsTTL).Err(); err != nil {
		log.Printf("cache items: store: %v", err)
	}
}
