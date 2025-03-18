package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kurochkinivan/Meet/internal/entity"
	"github.com/kurochkinivan/Meet/pkg/lfu"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type UserRepository struct {
	cache *lfu.LFUCache
}

func NewUserRepository(client *redis.Client, LFUCapacity int64, expiration time.Duration) *UserRepository {
	return &UserRepository{
		cache: lfu.New(client, LFUCapacity, expiration),
	}
}

func (r *UserRepository) Get(ctx context.Context, userID string) (*entity.User, bool) {
	key := fmt.Sprintf("user:%s", userID)
	userStr, err := r.cache.Get(ctx, key)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logrus.WithError(err).Errorf("failed to get user %q from cache", userID)
		}
		return nil, false
	}

	user := &entity.User{}
	err = json.Unmarshal([]byte(userStr), &user)
	if err != nil {
		logrus.WithError(err).Errorf("failed to unmarshall data for user %q", userID)
		return nil, false
	}

	return user, true
}

func (r *UserRepository) Set(ctx context.Context, user *entity.User) error {
	key := fmt.Sprintf("user:%s", user.UUID)
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal data for user %q: %w", user.UUID, err)
	}

	err = r.cache.Set(ctx, key, userData)
	if err != nil {
		return err
	}

	return nil
}
