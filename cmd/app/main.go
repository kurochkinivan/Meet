package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/kurochkinivan/Meet/config"
	"github.com/kurochkinivan/Meet/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logrus.Info("loading config")
	cfg := config.MustLoad()

	logrus.Info("setting up logger")
	setupLogger(cfg.Env)

	logrus.Info("initializing app")
	a, err := app.NewApp(ctx, cfg)
	if err != nil {
		logrus.WithError(err).Fatal("failed to initialize app")
	}

	logrus.Info("starting app")
	err = a.Run(ctx)
	if err != nil {
		logrus.WithError(err).Error("app.Run")
		return
	}
}

const (
	envDocker = "docker"
	envLocal  = "local"
	envProd   = "prod"
)

func setupLogger(env string) {
	callerPrettyfier := func(f *runtime.Frame) (string, string) {
		filename := path.Base(f.File)
		funcName := path.Base(f.Function)
		return fmt.Sprintf("%s()", funcName), fmt.Sprintf("%s:%d", filename, f.Line)
	}

	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)

	switch env {
	case envLocal, envDocker:
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:      true,
			TimestampFormat:  "2006-01-02 15:04:05",
			FullTimestamp:    true,
			CallerPrettyfier: callerPrettyfier,
		})
		logrus.SetLevel(logrus.TraceLevel)
	case envProd:
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:  "2006-01-02 15:04:05",
			CallerPrettyfier: callerPrettyfier,
		})
		logrus.SetLevel(logrus.InfoLevel)
	default:
		logrus.Fatal("unknown environment")
	}
}
