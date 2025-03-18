package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Env string `yaml:"env"`

	HTTP struct {
		Host         string        `yaml:"host" env:"HTTP_HOST" env-required:"true"`
		Port         string        `yaml:"port" env:"HTTP_PORT" env-required:"true"`
		ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-required:"true"`
		WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-required:"true"`
		IdleTimeout  time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-required:"true"`
		BytesLimit   int64         `yaml:"bytes_limit" env:"HTTP_BYTES_LIMIT" env-required:"true"`
		MaxLimit     int64         `yaml:"max_limit" env:"HTTP_MAX_LIMIT" env-required:"true"`
	} `yaml:"http"`

	PostgreSQL struct {
		Host     string `yaml:"host" env:"PSQL_HOST" env-required:"true"`
		Port     string `yaml:"port" env:"PSQL_PORT" env-required:"true"`
		Username string `yaml:"username" env:"PSQL_USERNAME" env-required:"true"`
		Password string `yaml:"password" env:"PSQL_PASSWORD" env-required:"true"`
		Database string `yaml:"database" env:"PSQL_DATABASE" env-required:"true"`
	} `yaml:"postgresql"`

	Redis struct {
		Host     string `yaml:"host" env:"REDIS_HOST" env-required:"true"`
		Port     string `yaml:"port" env:"REDIS_PORT" env-required:"true"`
		Password string `yaml:"password" env:"REDIS_PASSWORD" env-required:"true"`
		Database int    `yaml:"database" env:"REDIS_DATABASE" env-required:"true"`
	} `yaml:"redis"`

	S3 struct {
		BucketName string `yaml:"bucket_name" env:"S3_BUCKET_NAME" env-required:"true"`
	} `yaml:"s3"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		logrus.Warn("CONFIG_PATH is not set")
		configPath = "../../config/config.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logrus.WithError(err).Fatalf("CONFIG_PATH does not exist")
	}

	cfg := &Config{}
	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		logrus.WithError(err).Fatalf("Failed to read config")
	}

	return cfg
}
