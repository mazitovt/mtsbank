package repo

import (
	"context"
	"errors"
	api "mtsbank/history/internal/api/http/v1"
	"time"
)

var (
	ErrNoCurrencyPair = errors.New("currency pair doesn't exist in database")
)

type RegistryRow struct {
	CurrencyPair string
	Time         time.Time
	Rate         int64
}

// TODO: receive buffer to write query results
type Repo interface {
	Insert(ctx context.Context, data []RegistryRow) error
	InsertWithCurrencyPair(ctx context.Context, currencyPair string, data []api.ExchangeRate) error
	GetByTime(ctx context.Context, currencyPair string, start time.Time, end time.Time) ([]RegistryRow, error)
	Currencies(ctx context.Context) ([]string, error)
}
