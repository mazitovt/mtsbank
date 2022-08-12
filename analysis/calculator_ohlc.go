package analysis

import "mtsbank/analysis/internal/repo"

type CalculatorOHLC struct {
	ohlc  repo.OHLC
	empty bool
}

func NewCalculatorOHLC() CalculatorOHLC {
	return CalculatorOHLC{ohlc: repo.OHLC{}, empty: true}
}

func (c *CalculatorOHLC) Update(r ExchangeRate) {
	if c.empty {
		c.empty = false
		c.ohlc = repo.OHLC{
			OpenTime:  r.Time,
			CloseTime: r.Time,
			Open:      r.Rate,
			High:      r.Rate,
			Low:       r.Rate,
			Close:     r.Rate,
		}
		return
	}

	if r.Rate < c.ohlc.Low {
		c.ohlc.Low = r.Rate
	}
	if r.Rate > c.ohlc.High {
		c.ohlc.High = r.Rate
	}

	c.ohlc.Close = r.Rate
	c.ohlc.CloseTime = r.Time
}

func (c *CalculatorOHLC) OHLC() repo.OHLC {
	return c.ohlc
}

func (c *CalculatorOHLC) IsEmpty() bool {
	return c.empty
}

func (c *CalculatorOHLC) Reset() {
	c.ohlc = repo.OHLC{}
	c.empty = true
}
