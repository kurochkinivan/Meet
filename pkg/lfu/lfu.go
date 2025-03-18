package lfu

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	LFUSortedSetName = "LFUCacheSortedSet"
)

type LFUCache struct {
	redisClient *redis.Client
	expiration  time.Duration
	capacity    int64
}

func New(client *redis.Client, capacity int64, expiration time.Duration) *LFUCache {
	return &LFUCache{
		redisClient: client,
		expiration:  expiration,
		capacity:    capacity,
	}
}

func (c *LFUCache) Set(ctx context.Context, key string, value any) error {
	err := c.capacityCheck(ctx)
	if err != nil {
		return fmt.Errorf("failed to check capacity: %w", err)
	}

	err = c.redisClient.Set(ctx, key, value, c.expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to put key %q: %w", key, err)
	}

	err = c.IncrFrequency(ctx, key)
	if err != nil {
		return err
	}

	return nil
}

func (c *LFUCache) Get(ctx context.Context, key string) (string, error) {
	value, err := c.redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get value for key %q: %w", key, err)
	}

	err = c.IncrFrequency(ctx, key)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *LFUCache) Del(ctx context.Context, key string) error {
	err := c.capacityCheck(ctx)
	if err != nil {
		return fmt.Errorf("failed to check capacity: %w", err)
	}

	err = c.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %q: %w", key, err)
	}

	err = c.redisClient.ZRem(ctx, LFUSortedSetName, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %q from sorted set: %w", key, err)
	}

	return nil
}

func (c *LFUCache) Flush(ctx context.Context) error {
	err := c.redisClient.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush all data: %w", err)
	}

	return nil
}

func (c *LFUCache) IncrFrequency(ctx context.Context, key string) error {
	err := c.redisClient.ZIncrBy(ctx, LFUSortedSetName, 1, key).Err()
	if err != nil {
		return fmt.Errorf("failed to incr key %q in sorted set: %w", key, err)
	}

	return nil
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
			key, ok := item.Member.(string)
			if !ok {
				return fmt.Errorf("failed to assert key %q of type %[1]T to type string: %w", item.Member, err)
			}

			err = c.Del(ctx, key)
			if err != nil {
				return fmt.Errorf("failed to delete key %q: %w", key, err)
			}
		}
	}

	return nil
}
