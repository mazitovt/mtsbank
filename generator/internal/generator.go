package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"generator/internal/api/http/v1"
	"generator/pkg/cache"
	"github.com/mazitovt/logger"
	"net/http"
	"sync"
	"time"
)

var _ v1.ServerInterface = (*SimplePriceGenerator)(nil)

type SimplePriceGenerator struct {
	cache  map[string]cache.Cache[v1.ExchangeRate]
	f      GeneratorFunc
	logger logger.Logger
	pool   sync.Pool
}

func NewSimplePriceGenerator(currencyPairs []string, f GeneratorFunc, cacheSize uint64, logger logger.Logger) *SimplePriceGenerator {

	m := map[string]cache.Cache[v1.ExchangeRate]{}
	for _, p := range currencyPairs {
		m[p] = cache.NewLimitedCache[v1.ExchangeRate](cacheSize)
	}

	return &SimplePriceGenerator{
		cache:  m,
		f:      f,
		logger: logger,
		pool: sync.Pool{New: func() any {
			return make([]v1.ExchangeRate, 0, cacheSize)
		}},
	}
}

// Start generates new rates until context is Done
func (s *SimplePriceGenerator) Start(ctx context.Context, period time.Duration) {
	loggerLine := "SimplePriceGenerator.Start: "
	s.logger.Debug(loggerLine + "start")
	defer s.logger.Debug(loggerLine + "end")

	wg := sync.WaitGroup{}
	wg.Add(len(s.cache))
	for cur, c := range s.cache {
		go func(cur string, c cache.Cache[v1.ExchangeRate]) {
			defer wg.Done()
			s.generate(ctx, cur, c, period)
		}(cur, c)
	}
	wg.Wait()
	s.logger.Info("Generating stopped")
}

func (s *SimplePriceGenerator) GetRatesCurrencyPair(w http.ResponseWriter, r *http.Request, currencyPair string) {
	v, ok := s.cache[currencyPair]
	if !ok {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("service doesn't generate values for '%s'", currencyPair))
		return
	}

	// content is set to application/json only that order
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	out := s.pool.Get().([]v1.ExchangeRate)
	out = out[:0]
	defer s.pool.Put(out)

	out = v.Fill(out)

	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		s.logger.Error("Encode.Err: %v", err)
	}
}

func (s *SimplePriceGenerator) generate(ctx context.Context, cur string, cache cache.Cache[v1.ExchangeRate], period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		t := time.Now().Round(time.Microsecond)
		rate := s.f(cur)
		exRate := v1.ExchangeRate{
			Time: t,
			Rate: rate,
		}
		s.logger.Debug("currency=%v, rate=%v", cur, exRate)
		cache.Put(exRate)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			continue
		}
	}
}

func (s *SimplePriceGenerator) writeError(w http.ResponseWriter, code int, message string) {
	petErr := v1.Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(petErr)
	if err != nil {
		s.logger.Error("Encode.Err: %v", err)
	}
}
