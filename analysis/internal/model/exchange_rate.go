package model

import "time"

type ExchangeRate struct {
	Time time.Time
	Rate int64
}
