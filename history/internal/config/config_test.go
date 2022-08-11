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
				"RATE_HISTORY_HOST":    "127.0.0.1",
				"RATE_HISTORY_PORT":    "8081",
				"RATE_HISTORY_MIGRATE": "true",
				"RATE_HISTORY_PERIOD":  "5s",

				"RATE_HISTORY_GENERATOR_HOST": "generator",
				"RATE_HISTORY_GENERATOR_PORT": "8080",

				"RATE_HISTORY_POSTGRES_HOST":     "postgres",
				"RATE_HISTORY_POSTGRES_PORT":     "5432",
				"RATE_HISTORY_POSTGRES_USER":     "history",
				"RATE_HISTORY_POSTGRES_PASSWORD": "history",
				"RATE_HISTORY_POSTGRES_SSLMODE":  "disable",
				"RATE_HISTORY_POSTGRES_DBNAME":   "history",
			},
			er: Config{
				Host:    "127.0.0.1",
				Port:    "8080",
				Migrate: true,
				Period:  5 * time.Second,
				Generator: Generator{
					Host: "generator",
					Port: "8080",
				},
				Postgres: PostgresConfig{
					Host:     "postgres",
					Port:     "5432",
					User:     "history",
					Password: "history",
					DBname:   "history",
					Sslmode:  "disable",
				},
			},
		},
		{
			name: "period: 0.5s",
			inputEnv: map[string]string{
				"RATE_HISTORY_HOST":    "127.0.0.1",
				"RATE_HISTORY_PORT":    "8080",
				"RATE_HISTORY_MIGRATE": "true",
				"RATE_HISTORY_PERIOD":  "0.5s",

				"RATE_HISTORY_GENERATOR_HOST": "generator",
				"RATE_HISTORY_GENERATOR_PORT": "8080",

				"RATE_HISTORY_POSTGRES_HOST":     "postgres",
				"RATE_HISTORY_POSTGRES_PORT":     "5432",
				"RATE_HISTORY_POSTGRES_USER":     "history",
				"RATE_HISTORY_POSTGRES_PASSWORD": "history",
				"RATE_HISTORY_POSTGRES_SSLMODE":  "disable",
				"RATE_HISTORY_POSTGRES_DBNAME":   "history",
			},
			er: Config{
				Host:    "127.0.0.1",
				Port:    "8080",
				Migrate: true,
				Period:  5 * time.Second,
				Generator: Generator{
					Host: "generator",
					Port: "8080",
				},
				Postgres: PostgresConfig{
					Host:     "postgres",
					Port:     "5432",
					User:     "history",
					Password: "history",
					DBname:   "history",
					Sslmode:  "disable",
				},
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
