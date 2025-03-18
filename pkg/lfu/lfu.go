package lfu

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	LFUHashName      = "LFUCacheHash"
	LFUSortedSetName = "LFUCacheSortedSet"
)

type LFUCache struct {
	redisClient *redis.Client
	capacity    int64
}

func New(client *redis.Client, capacity int64) *LFUCache {
	return &LFUCache{
		redisClient: client,
		capacity:    capacity,
	}
}

func (c *LFUCache) Put(ctx context.Context, key string, value any) error {
	c.capacityCheck()

	err := c.redisClient.HSet(ctx, LFUHashName, key, value).Err()
	if err != nil {
		return fmt.Errorf("failed to put key %q into LFU cache: %w", key, err)
	}

	go c.frequentlyIncr(ctx, key)

	return nil
}

func (c *LFUCache) Del(ctx context.Context, key string) error {
	c.capacityCheck()

	err := c.redisClient.HDel(ctx, LFUHashName, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %q from LFU cache: %w", key, err)
	}

	err = c.redisClient.ZRem(ctx, LFUSortedSetName, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %q from LFU sorted set: %w", key, err)
	}

	return nil
}

func (c *LFUCache) frequentlyIncr(ctx context.Context, key string) {
	c.redisClient.ZIncrBy(ctx, LFUSortedSetName, 1, key)
}

func (c *LFUCache) capacityCheck(ctx context.Context) error {
	capacity, _ := c.redisClient.ZCard(ctx, LFUSortedSetName).Result()
	if capacity >= c.capacity {
		deleteCount := capacity - c.capacity + 1
		items, err := c.redisClient.ZPopMin(ctx, LFUSortedSetName, deleteCount).Result()
		if err != nil {
			return fmt.Errorf("failed to pop min from sorted set: %w", err)
		}

		for _, item := range items {
			c.Del(ctx, item.Member.(string))
		}
	}
}
