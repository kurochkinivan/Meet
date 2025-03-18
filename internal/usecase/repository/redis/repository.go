package redis

import "github.com/redis/go-redis/v9"

type Repositories struct {
}

func NewRepositories(client *redis.Client) *Repositories {
	return &Repositories{}
}
