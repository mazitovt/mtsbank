package app

import (
	"fmt"
	"generator/app/cache"
	"generator/app/logger"
	"math/rand"
	"time"
)

type PriceGenerator interface {
	Get(currencyPair string) ([]ExchangeRateAt, error)
	Generate()
}

type GeneratorFunc func(string) int64

type ExchangeRateAt struct {
	Time         time.Time `json:"time"`
	ExchangeRate int64     `json:"exchangeRate"`
}

type SimplePriceGenerator struct {
	cache  map[string]cache.Cache[ExchangeRateAt]
	f      GeneratorFunc
	logger logger.Logger
}

func NewSimplePriceGenerator(currencyPairs []string, f GeneratorFunc, cacheSize uint64, logger logger.Logger) *SimplePriceGenerator {

	m := map[string]cache.Cache[ExchangeRateAt]{}
	for _, p := range currencyPairs {
		m[p] = cache.NewLimitedCache[ExchangeRateAt](cacheSize)
	}

	return &SimplePriceGenerator{cache: m, f: f, logger: logger}
}

func (s *SimplePriceGenerator) Get(currencyPair string) ([]ExchangeRateAt, error) {
	v, ok := s.cache[currencyPair]
	if !ok {
		return nil, fmt.Errorf("no such currency pair: %s", currencyPair)
	}
	return v.Values(), nil
}

func (s *SimplePriceGenerator) Generate() {
	t := time.Now().Round(time.Microsecond)
	for k, v := range s.cache {
		rate := s.f(k)
		exRate := ExchangeRateAt{
			Time:         t,
			ExchangeRate: rate,
		}
		s.logger.Debug("currency=%v, rate=%v", k, exRate)
		v.Add(exRate)
	}
}

func NewExchangeRateFromSeed(seed int64) GeneratorFunc {
	r := rand.New(rand.NewSource(seed))
	return func(currencyPair string) int64 {
		b := []byte(currencyPair)
		s := int64(0)
		for _, v := range b {
			s += int64(v)
		}
		return r.Int63() % s
	}
}

func ExchangeRateFromTime(currencyPair string) int64 {
	b := []byte(currencyPair)
	s := int64(0)
	for _, v := range b {
		s += int64(v)
	}

	return rand.New(rand.NewSource(time.Now().UnixNano())).Int63() % s
}
