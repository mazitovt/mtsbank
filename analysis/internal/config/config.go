package config

import (
	"errors"
	"github.com/kelseyhightower/envconfig"
	"time"
)

const (
	envPrefix = "RATE_ANALYZER"
)

var (
	ErrMinimalPeriod = errors.New("PERIOD must be equal or greater than 1 second (1s)")
)

type (
	Config struct {
		LogLevel      string          `envconfig:"LOG_LEVEL"`
		Batch         Batch           `envconfig:"BATCH"`
		Http          Http            `envconfig:"HTTP"`
		CurrencyPairs []string        `envconfig:"CURRENCY_PAIRS"`
		TimeFrames    []time.Duration `envconfig:"TIME_FRAMES"`
		PollPeriod    time.Duration   `envconfig:"POLL_PERIOD"`
		RestartAfter  time.Duration   `envconfig:"RESTART_AFTER"`
		History       HttpService     `envconfig:"HISTORY"`
		Generator     HttpService     `envconfig:"GENERATOR"`
	}

	Batch struct {
		Period time.Duration `envconfig:"PERIOD"`
		Size   int64         `envconfig:"SIZE"`
	}

	Http struct {
		Host string `envconfig:"HOST"`
		Port string `envconfig:"PORT"`
	}

	HttpService struct {
		Host string `envconfig:"HOST"`
		Port string `envconfig:"PORT"`
	}
)

func Init() (*Config, error) {
	cfg := &Config{}

	if err := envconfig.Process(envPrefix, cfg); err != nil {
		return nil, err
	}

	//TODO: dont return error, continue execution with period = 1s
	if cfg.PollPeriod < time.Second {
		return nil, ErrMinimalPeriod
	}

	return cfg, nil
}
