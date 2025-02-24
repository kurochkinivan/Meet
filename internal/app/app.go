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
	"github.com/kurochkinivan/Meet/internal/usecase/repository/postgresql"
	"github.com/kurochkinivan/Meet/pkg/psql"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type App struct {
	cfg    *config.Config
	server *http.Server
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	cfgpq := cfg.PostgreSQL
	pgConfig := psql.NewPgConfig(cfgpq.Username, cfgpq.Password, cfgpq.Host, cfgpq.Port, cfgpq.Database)

	logrus.Info("connecting to database client...")
	client, err := psql.NewClient(context.Background(), 5, pgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repositories := postgresql.NewRepositories(client)
	usecases := usecase.NewUseCases(repositories)
	handler := v1.NewHandler(usecases, cfg.HTTP.BytesLimit)

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
