package internal

import (
	"math/rand"
	"time"
)

type GeneratorFunc func(string) int64

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
