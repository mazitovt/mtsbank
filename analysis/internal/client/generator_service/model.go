package generator_service

import "time"

type ExchangeRateDTO struct {
	Time time.Time `json:"time"`
	Rate int64     `json:"exchangeRate"`
}
