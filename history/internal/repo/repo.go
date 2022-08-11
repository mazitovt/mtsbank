package repo

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNoCurrencyPair = errors.New("currency pair isn't present in database")
)

type RegistryRow struct {
	CurrencyPair string
	Time         time.Time
	Rate         int64
}

type Repo interface {
	Insert(ctx context.Context, data []RegistryRow) error
	GetByTime(ctx context.Context, currencyPair string, start time.Time, end time.Time) ([]RegistryRow, error)
	Currencies(ctx context.Context) ([]string, error)
}
