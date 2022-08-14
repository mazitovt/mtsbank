package repo

import (
	"context"
	"mtsbank/analysis/internal/model"
	"mtsbank/analysis/logger"
	"sync"
)

type Repo interface {
	Put(ctx context.Context, currencyPair string, timeFrame string, ohlc model.OHLC) error
	GetLast(ctx context.Context, currencyPair string, timeFrame string) (*model.OHLC, error)
	PutMany(ctx context.Context, ohlcs []model.OHLC) error

	// If last is bigger the number of existing OHLC, than return all OHLC
	GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64) ([]model.OHLC, error)
	Reset(ctx context.Context)
}

type InmemoryRepo struct {
	mu     sync.RWMutex
	m      map[string]map[string][]model.OHLC
	logger logger.Logger
}

func (r *InmemoryRepo) PutMany(ctx context.Context, ohlcs []model.OHLC) error {
	r.logger.Info("InmemoryRepo.PutMany: start: %v", len(ohlcs))
	defer r.logger.Info("InmemoryRepo.PutMany: end: %v", len(ohlcs))
	for _, ohlc := range ohlcs {
		_ = r.Put(ctx, ohlc.CurrencyPair, ohlc.TimeFrame.String(), ohlc)
	}
	return nil
}

func NewInmemoryRepo(logger logger.Logger) *InmemoryRepo {
	return &InmemoryRepo{m: map[string]map[string][]model.OHLC{}, logger: logger}
}

func (r *InmemoryRepo) Put(ctx context.Context, currencyPair string, timeFrame string, ohlc model.OHLC) error {
	r.logger.Debug("InmemoryRepo.Put: %v %v %v", currencyPair, timeFrame, ohlc)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.m[currencyPair]; !ok {
		r.m[currencyPair] = map[string][]model.OHLC{}
	}

	if _, ok := r.m[currencyPair][timeFrame]; !ok {
		r.m[currencyPair][timeFrame] = []model.OHLC{}
	}

	r.m[currencyPair][timeFrame] = append(r.m[currencyPair][timeFrame], ohlc)

	return nil
}

func (r *InmemoryRepo) GetLast(ctx context.Context, currencyPair string, timeFrame string) (*model.OHLC, error) {
	//TODO implement me
	panic("implement me")
}

func (r *InmemoryRepo) GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64) ([]model.OHLC, error) {
	//TODO implement me
	panic("implement me")
}

func (r *InmemoryRepo) Reset(ctx context.Context) {
	//r.m = map[string]map[string][]model.OHLC{}
}

// TODO: return copy
func (r *InmemoryRepo) GetAll() map[string]map[string][]model.OHLC {
	return r.m
}
