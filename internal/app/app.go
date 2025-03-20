package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/kurochkinivan/Meet/config"
	v1 "github.com/kurochkinivan/Meet/internal/controller/http/v1"
	"github.com/kurochkinivan/Meet/internal/usecase"
	"github.com/kurochkinivan/Meet/internal/usecase/repository/pg"
	"github.com/kurochkinivan/Meet/internal/usecase/repository/redis"
	"github.com/kurochkinivan/Meet/internal/usecase/repository/s3"
	pgclient "github.com/kurochkinivan/Meet/pkg/pgClient"
	redisclient "github.com/kurochkinivan/Meet/pkg/redisClient"
	s3client "github.com/kurochkinivan/Meet/pkg/s3Client"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type App struct {
	cfg    *config.Config
	server *http.Server
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	logrus.Info("connecting to postgresql...")
	clientPSQL, err := pgclient.NewClient(context.Background(), 5, &pgclient.PgConfig{
		Username: cfg.PostgreSQL.Username,
		Password: cfg.PostgreSQL.Password,
		Host:     cfg.PostgreSQL.Host,
		Port:     cfg.PostgreSQL.Port,
		Database: cfg.PostgreSQL.Database,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgresql: %w", err)
	}

	logrus.Info("connecting to redis...")
	clientRedis, err := redisclient.NewClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, int(*cfg.Redis.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logrus.Info("connecting to s3...")
	clientS3, err := s3client.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client: %w", err)
	}

	pgRepositories := pg.NewRepositories(clientPSQL)
	redisRepositories := redis.NewRepositories(clientRedis, cfg.Redis.LFUCapacity, cfg.Redis.Expiration)
	s3Repositories := s3.NewRepositories(clientS3, cfg.S3.BucketName)

	usecases := usecase.NewUseCases(cfg, pgRepositories, s3Repositories, redisRepositories)

	handler := v1.NewHandler(usecases, cfg.HTTP.BytesLimit, cfg.HTTP.MaxLimit)

	server := &http.Server{
		Addr:         net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      handler,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &App{
		server: server,
		cfg:    cfg,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	grp, ctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return a.startHTTP(ctx)
	})

	return grp.Wait()
}

func (a *App) startHTTP(ctx context.Context) error {
	err := a.server.ListenAndServe()
	if err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logrus.Warn("server shutdown")
		default:
			logrus.WithError(err).Fatal("failed to start server")
		}
	}

	err = a.server.Shutdown(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("failed to shutdown server")
	}

	return err
}
