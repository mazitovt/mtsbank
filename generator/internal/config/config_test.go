package config

import (
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/env"
	"testing"
	"time"
)

func TestConfig_Init(t *testing.T) {
	tests := []struct {
		name     string
		inputEnv map[string]string
		er       Config
		err      error
	}{
		{
			name: "config with time",
			inputEnv: map[string]string{
				"RATE_GENERATOR_CURRENCY_PAIRS": "EURUSD,USDRUB,USDJPY",
				"RATE_GENERATOR_PATTERN":        "TIME",
				"RATE_GENERATOR_PERIOD":         "1s",
				"RATE_GENERATOR_CACHE_SIZE":     "5",
			},
			er: Config{
				CurrencyPairs: []string{"EURUSD", "USDRUB", "USDJPY"},
				Pattern:       "TIME",
				Period:        1 * time.Second,
				CacheSize:     5,
			},
		},
		{
			name: "config with seed",
			inputEnv: map[string]string{
				"RATE_GENERATOR_CURRENCY_PAIRS": "EURUSD,USDRUB,USDJPY",
				"RATE_GENERATOR_PATTERN":        "SEED",
				"RATE_GENERATOR_SEED":           "123",
				"RATE_GENERATOR_PERIOD":         "3s",
				"RATE_GENERATOR_CACHE_SIZE":     "8",
			},
			er: Config{
				CurrencyPairs: []string{"EURUSD", "USDRUB", "USDJPY"},
				Pattern:       "SEED",
				Seed:          123,
				Period:        3 * time.Second,
				CacheSize:     8,
			},
		},
		{
			name: "config missing seed",
			inputEnv: map[string]string{
				"RATE_GENERATOR_CURRENCY_PAIRS": "EURUSD,USDRUB,USDJPY",
				"RATE_GENERATOR_PATTERN":        "SEED",
				"RATE_GENERATOR_PERIOD":         "3s",
				"RATE_GENERATOR_CACHE_SIZE":     "8",
			},
			er: Config{
				CurrencyPairs: []string{"EURUSD", "USDRUB", "USDJPY"},
				Pattern:       "SEED",
				Period:        3 * time.Second,
				CacheSize:     8,
			},
		},
		{
			name: "config negative cached size",
			inputEnv: map[string]string{
				"RATE_GENERATOR_CURRENCY_PAIRS": "EURUSD,USDRUB,USDJPY",
				"RATE_GENERATOR_PATTERN":        "TIME",
				"RATE_GENERATOR_PERIOD":         "3s",
				"RATE_GENERATOR_CACHE_SIZE":     "-8",
			},
			err: ErrMinimalCacheSize,
		},
		{
			name: "100 milisecond period",
			inputEnv: map[string]string{
				"RATE_GENERATOR_CURRENCY_PAIRS": "EURUSD,USDRUB,USDJPY",
				"RATE_GENERATOR_PATTERN":        "TIME",
				"RATE_GENERATOR_PERIOD":         "0.5s",
				"RATE_GENERATOR_CACHE_SIZE":     "5",
			},
			err: ErrMinimalPeriod,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer env.PatchAll(t, tc.inputEnv)()
			cfg, err := Init()
			if tc.err != nil {
				require.Equal(t, tc.err, err)
				return
			}
			require.Nil(t, err)
			if !cmp.Equal(tc.er, *cfg) {
				t.Fatal("configs aren't equal\n", cmp.Diff(tc.er, *cfg))
			}
		})
	}
}
