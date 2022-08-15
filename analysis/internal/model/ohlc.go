package model

import "time"

var (
	emptyTime = time.Time{}
	emptyRate = int64(0)
)

type OHLC struct {
	CurrencyPair string
	TimeFrame    time.Duration
	OpenTime     time.Time
	CloseTime    time.Time
	Open         int64
	High         int64
	Low          int64
	Close        int64
}

func NewOHLC(currencyPair string, timeFrame time.Duration) *OHLC {
	return &OHLC{CurrencyPair: currencyPair, TimeFrame: timeFrame}
}

func (o *OHLC) UpdateOrReady(r ExchangeRate) bool {
	if o.empty() {
		o.OpenTime = r.Time
		o.CloseTime = r.Time
		o.Open = r.Rate
		o.High = r.Rate
		o.Low = r.Rate
		o.Close = r.Rate
		return false
	}

	if r.Time.Sub(o.OpenTime) > o.TimeFrame {
		return true
	}

	if r.Rate < o.Low {
		o.Low = r.Rate
	}
	if r.Rate > o.High {
		o.High = r.Rate
	}

	o.Close = r.Rate
	o.CloseTime = r.Time

	return false
}

func (o *OHLC) Empty() bool {
	return o.OpenTime == emptyTime
}

func (o *OHLC) Reset() {
	o.OpenTime = emptyTime
	o.CloseTime = emptyTime
	o.Open = emptyRate
	o.High = emptyRate
	o.Low = emptyRate
	o.Close = emptyRate
}

func (o *OHLC) empty() bool {
	return o.OpenTime == emptyTime
}
