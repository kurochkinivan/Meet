package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Repositories struct {
	*UserRepository
}

// TODO: remove hardcode
func NewRepositories(client *redis.Client, LFUCapacity int64, expiration time.Duration) *Repositories {
	return &Repositories{
		UserRepository: NewUserRepository(client, LFUCapacity, expiration),
	}
}
