package redisclient

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func NewClient(host, port, password string, database int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(host, port),
		Password: password,
		DB:       database,
	})

	err := doWithAttempts(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		status := client.Ping(ctx)
		if status.Err() != nil {
			logrus.Warn("failed to connect to redis, retrying...")
			return status.Err()
		}

		return nil
	}, 5*time.Second, 5)
	if err != nil {
		return nil, fmt.Errorf("all attempts exceeded, failed to connect to redis, err: %w", err)
	}

	return client, nil
}

func doWithAttempts(f func() error, timeout time.Duration, maxAttempts int) error {
	var err error
	for maxAttempts > 0 {
		if err = f(); err != nil {
			time.Sleep(timeout)
			maxAttempts--
			continue
		}
		return nil
	}
	return err
}
