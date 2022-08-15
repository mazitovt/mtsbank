package config

import (
	"errors"
	"github.com/kelseyhightower/envconfig"
	"time"
)

const (
	envPrefix = "RATE_HISTORY"
)

var (
	ErrMinimalPeriod = errors.New("PERIOD must be equal or greater than 1 second (1s)")
)

type (
	Config struct {
		LogLevel  string         `envconfig:"LOG_LEVEL"`
		Host      string         `envconfig:"HOST"`
		Port      string         `envconfig:"PORT"`
		Period    time.Duration  `envconfig:"PERIOD"`
		Migrate   bool           `envconfig:"MIGRATE"`
		Postgres  PostgresConfig `envconfig:"POSTGRES"`
		Generator Generator      `envconfig:"GENERATOR"`
	}

	Generator struct {
		Host string `envconfig:"HOST"`
		Port string `envconfig:"PORT"`
	}

	PostgresConfig struct {
		Host     string `envconfig:"HOST"`
		Port     string `envconfig:"PORT"`
		User     string `envconfig:"USER"`
		Password string `envconfig:"PASSWORD"`
		DBname   string `envconfig:"DBNAME"`
		Sslmode  string `envconfig:"SSLMODE"`
	}
)

func Init() (*Config, error) {
	cfg := &Config{}

	if err := envconfig.Process(envPrefix, cfg); err != nil {
		return nil, err
	}

	if cfg.Period < time.Second {
		return nil, ErrMinimalPeriod
	}

	return cfg, nil
}
