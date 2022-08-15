package config

import (
	"errors"
	"fmt"
	"generator/internal"
	"time"
)
import "github.com/kelseyhightower/envconfig"

const (
	envPrefix = "RATE_GENERATOR"
)

var (
	ErrMinimalCacheSize = errors.New("CACHE_SIZE must be equal or greater than zero")
	ErrMinimalPeriod    = errors.New("PERIOD must be equal or greater than 1 second (1s)")
)

type Config struct {
	LogLevel      string        `envconfig:"LOG_LEVEL"`
	Host          string        `envconfig:"HOST"`
	Port          string        `envconfig:"PORT"`
	CurrencyPairs []string      `envconfig:"CURRENCY_PAIRS"`
	Pattern       string        `envconfig:"PATTERN"`
	Seed          int64         `envconfig:"SEED"`
	Period        time.Duration `envconfig:"PERIOD"`
	CacheSize     int64         `envconfig:"CACHE_SIZE"`
}

func Init() (*Config, error) {
	cfg := &Config{}

	if err := envconfig.Process(envPrefix, cfg); err != nil {
		return nil, err
	}

	if cfg.CacheSize < 0 {
		return nil, ErrMinimalCacheSize
	}

	//TODO: dont return error, continue execution with period = 1s
	if cfg.Period < time.Second {
		return nil, ErrMinimalPeriod
	}

	return cfg, nil
}

func GetGeneratorFunc(cfg *Config) (internal.GeneratorFunc, error) {
	switch cfg.Pattern {
	case "TIME":
		return internal.ExchangeRateFromTime, nil
	case "SEED":
		return internal.NewExchangeRateFromSeed(cfg.Seed), nil
	}
	return nil, fmt.Errorf("unknown pattern: %s", cfg.Pattern)
}
