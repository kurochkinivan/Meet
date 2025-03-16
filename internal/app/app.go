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
	"github.com/kurochkinivan/Meet/internal/usecase/repository/s3"
	pgclient "github.com/kurochkinivan/Meet/pkg/pgClient"
	s3client "github.com/kurochkinivan/Meet/pkg/s3Client"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type App struct {
	cfg    *config.Config
	server *http.Server
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	cfgpq := cfg.PostgreSQL
	pgConfig := pgclient.NewPgConfig(cfgpq.Username, cfgpq.Password, cfgpq.Host, cfgpq.Port, cfgpq.Database)

	logrus.Info("connecting to database client...")
	clientPSQL, err := pgclient.NewClient(context.Background(), 5, pgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	clientS3, err := s3client.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client: %w", err)
	}

	pgRepositories := pg.NewRepositories(clientPSQL)
	s3Repositories := s3.NewRepositories(clientS3, cfg.S3.BucketName)
	usecases := usecase.NewUseCases(pgRepositories, s3Repositories)
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
