package repo

import (
	"context"
	"mtsbank/analysis/logger"
	"time"
)

type OHLC struct {
	CurrencyPair string
	TimeFrame    string
	OpenTime     time.Time
	CloseTime    time.Time
	Open         int64
	High         int64
	Low          int64
	Close        int64
}

func (o *OHLC) Init() {

}

func (o *OHLC) Update() {

}

type Repo interface {
	Put(ctx context.Context, currencyPair string, timeFrame string, ohlc OHLC) error
	GetLast(ctx context.Context, currencyPair string, timeFrame string) (*OHLC, error)

	// If last is bigger the number of existing OHLC, than return all OHLC
	GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64) ([]OHLC, error)
	Reset(ctx context.Context)
}

type InmemoryRepo struct {
	m      map[string]map[string][]OHLC
	logger logger.Logger
}

func NewInmemoryRepo(logger logger.Logger) *InmemoryRepo {
	return &InmemoryRepo{m: map[string]map[string][]OHLC{}, logger: logger}
}

func (r *InmemoryRepo) Put(ctx context.Context, currencyPair string, timeFrame string, ohlc OHLC) error {
	r.logger.Debug("InmemoryRepo.Put: %v %v %v", currencyPair, timeFrame, ohlc)
	if _, ok := r.m[currencyPair]; !ok {
		r.m[currencyPair] = map[string][]OHLC{}
	}

	if _, ok := r.m[currencyPair][timeFrame]; !ok {
		r.m[currencyPair][timeFrame] = []OHLC{}
	}

	r.m[currencyPair][timeFrame] = append(r.m[currencyPair][timeFrame], ohlc)

	return nil
}

func (r *InmemoryRepo) GetLast(ctx context.Context, currencyPair string, timeFrame string) (*OHLC, error) {
	//TODO implement me
	panic("implement me")
}

func (r *InmemoryRepo) GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64) ([]OHLC, error) {
	//TODO implement me
	panic("implement me")
}

func (r *InmemoryRepo) Reset(ctx context.Context) {
	r.m = map[string]map[string][]OHLC{}
}
