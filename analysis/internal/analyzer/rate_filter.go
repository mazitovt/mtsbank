package analyzer

import (
	"mtsbank/analysis/internal/model"
	"time"
)

type rateFilter struct {
	last time.Time
}

// Check takes ORDERED rates and return index, from which starts unseen data
func (t *rateFilter) Check(rates []model.ExchangeRate) int64 {
	for i, r := range rates {
		if r.Time.After(t.last) {
			t.last = rates[len(rates)-1].Time
			return int64(i)
		}
	}
	return int64(len(rates))
}

func (t *rateFilter) Last() time.Time {
	return t.last
}
