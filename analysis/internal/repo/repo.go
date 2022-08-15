package repo

import (
	"context"
	"errors"
	"github.com/mazitovt/logger"
	"mtsbank/analysis/internal/model"
	"sync"
	"time"
)

var (
	ErrBadCurrencyPairOrTimeFrame = errors.New("currency pair or time frame doesn't exit in database")
	ErrBadCurrencyPair            = errors.New("currency pair doesn't exist in database")
	ErrBadTimeFrame               = errors.New("time frame doesn't exist for currency pair in database")
)

type Repo interface {
	Put(ctx context.Context, currencyPair string, timeFrame string, ohlc model.OHLC) error
	PutMany(ctx context.Context, ohlcs []model.OHLC) error

	GetLast(ctx context.Context, currencyPair string, timeFrame string) (*model.OHLC, error)
	GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64, buffer []model.OHLC) ([]model.OHLC, error)
	GetManyFromTo(ctx context.Context, currencyPair string, timeFrame string, from time.Time, to time.Time, buffer []model.OHLC) ([]model.OHLC, error)
	Reset(ctx context.Context)
}

var _ Repo = (*InmemoryRepo)(nil)

type InmemoryRepo struct {
	mu     sync.RWMutex
	m      map[string][]model.OHLC
	logger logger.Logger
}

func NewInmemoryRepo(logger logger.Logger) *InmemoryRepo {
	return &InmemoryRepo{m: map[string][]model.OHLC{}, logger: logger}
}

func (r *InmemoryRepo) Put(ctx context.Context, currencyPair string, timeFrame string, ohlc model.OHLC) error {
	r.logger.Info("InmemoryRepo.Put: '%v' '%v' '%v'", currencyPair, timeFrame, ohlc)

	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := currencyPair + timeFrame
	if _, ok := r.m[key]; !ok {
		r.m[key] = []model.OHLC{}
	}

	r.m[key] = append(r.m[key], ohlc)

	return nil
}

func (r *InmemoryRepo) PutMany(ctx context.Context, ohlcs []model.OHLC) error {
	r.logger.Debug("InmemoryRepo.PutMany: start: %v", len(ohlcs))
	defer r.logger.Debug("InmemoryRepo.PutMany: end: %v", len(ohlcs))
	r.logger.Info("InmemoryRepo.PutMany: ohlcs: %v", ohlcs)
	for _, ohlc := range ohlcs {
		_ = r.Put(ctx, ohlc.CurrencyPair, ohlc.TimeFrame.String(), ohlc)
	}
	return nil
}

func (r *InmemoryRepo) GetLast(ctx context.Context, currencyPair string, timeFrame string) (*model.OHLC, error) {
	r.logger.Debug("InmemoryRepo.GetLast: %v %v", currencyPair, timeFrame)

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := currencyPair + timeFrame

	ohlcs, ok := r.m[key]
	if !ok {
		return nil, ErrBadCurrencyPairOrTimeFrame
	}

	o := ohlcs[len(ohlcs)-1]

	return &o, nil
}

// TODO: implement
func (r *InmemoryRepo) GetManyFromTo(ctx context.Context, currencyPair string, timeFrame string, from time.Time, to time.Time, buffer []model.OHLC) ([]model.OHLC, error) {
	return r.getAll(ctx, currencyPair, timeFrame, buffer)
}

func (r *InmemoryRepo) GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64, buffer []model.OHLC) ([]model.OHLC, error) {
	buffer, err := r.getAll(ctx, currencyPair, timeFrame, buffer)
	if err != nil {
		return buffer, err
	}

	len64 := int64(len(buffer))
	if last > len64 {
		last = len64
	}

	return buffer[len64-last:], nil
}

func (r *InmemoryRepo) getAll(ctx context.Context, currencyPair string, timeFrame string, buffer []model.OHLC) ([]model.OHLC, error) {
	r.logger.Debug("InmemoryRepo.GetLast: %v %v", currencyPair, timeFrame)

	select {
	default:
	case <-ctx.Done():
		return buffer, ctx.Err()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := currencyPair + timeFrame

	ohlcs, ok := r.m[key]
	if !ok {
		return buffer, ErrBadCurrencyPairOrTimeFrame
	}

	if len(ohlcs) > cap(buffer) {
		buffer = make([]model.OHLC, 0, len(ohlcs))
	}

	buffer = buffer[:len(ohlcs)]

	copy(buffer, ohlcs)

	return buffer, nil
}

func (r *InmemoryRepo) Reset(ctx context.Context) {
	//r.mu.Lock()
	//defer r.mu.Unlock()
	//r.m = map[string][]model.OHLC{}
}

// TODO: return copy
func (r *InmemoryRepo) GetMap() map[string][]model.OHLC {
	return r.m
}
