package repo

import (
	"context"
	"time"
)

type OHLC struct {
	OpenTime  time.Time
	CloseTime time.Time
	Open      int64
	High      int64
	Low       int64
	Close     int64
}
type Repo interface {
	Put(ctx context.Context, currencyPair string, timeFrame string, ohlc OHLC) error
	Get(ctx context.Context, currencyPair string, timeFrame string) (*OHLC, error)

	// If last is bigger the number of existing OHLC, than return all OHLC
	GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64) ([]OHLC, error)
	Reset(ctx context.Context)
}
