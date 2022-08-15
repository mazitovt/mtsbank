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
			name: "valid config",
			inputEnv: map[string]string{
				"RATE_ANALYZER_LOG_LEVEL":      "debug",
				"RATE_ANALYZER_BATCH_PERIOD":   "5s",
				"RATE_ANALYZER_BATCH_SIZE":     "50",
				"RATE_ANALYZER_HTTP_HOST":      "0.0.0.0",
				"RATE_ANALYZER_HTTP_PORT":      "8082",
				"RATE_ANALYZER_CURRENCY_PAIRS": "EURUSD,USDRUB,USDJPY",
				"RATE_ANALYZER_TIME_FRAMES":    "1m,5m,15m,30m,1h",
				"RATE_ANALYZER_POLL_PERIOD":    "1s",
				"RATE_ANALYZER_RESTART_AFTER":  "24h",
				"RATE_ANALYZER_HISTORY_HOST":   "localhost",
				"RATE_ANALYZER_HISTORY_PORT":   "8081",
				"RATE_ANALYZER_GENERATOR_HOST": "localhost",
				"RATE_ANALYZER_GENERATOR_PORT": "8080",
			},
			er: Config{
				Batch: Batch{
					Period: 5 * time.Second,
					Size:   50,
				},
				LogLevel: "debug",
				Http: Http{
					Host: "0.0.0.0",
					Port: "8082",
				},
				CurrencyPairs: []string{"EURUSD", "USDRUB", "USDJPY"},
				TimeFrames:    []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute, 30 * time.Minute, 1 * time.Hour},
				PollPeriod:    1 * time.Second,
				RestartAfter:  24 * time.Hour,
				Generator: HttpService{
					Host: "localhost",
					Port: "8080",
				},
				History: HttpService{
					Host: "localhost",
					Port: "8081",
				},
			},
		},
		{
			name: "period less than minimal value",
			inputEnv: map[string]string{
				"RATE_ANALYZER_HTTP_HOST":      "localhost",
				"RATE_ANALYZER_HTTP_PORT":      "8082",
				"RATE_ANALYZER_CURRENCY_PAIRS": "EURUSD,USDRUB,USDJPY",
				"RATE_ANALYZER_TIME_FRAMES":    "1m,5m,15m,30m,1h",
				"RATE_ANALYZER_POLL_PERIOD":    "0.5s",
				"RATE_ANALYZER_RESTART_AFTER":  "24h",
				"RATE_ANALYZER_HISTORY_HOST":   "localhost",
				"RATE_ANALYZER_HISTORY_PORT":   "8081",
				"RATE_ANALYZER_GENERATOR_HOST": "localhost",
				"RATE_ANALYZER_GENERATOR_PORT": "8080",
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
