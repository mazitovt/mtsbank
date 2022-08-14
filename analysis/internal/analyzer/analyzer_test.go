package analyzer

import (
	"fmt"
	"mtsbank/analysis/internal/model"
	"testing"
	"time"
)

func TestRateFilter(t *testing.T) {

	//2022-08-13 19:59:29.456658857 +0500 +05
	//2022-08-13 19:59:24.452834596 +0500 +05
	//2022-08-13 19:59:24.452835 +0000 UTC

	r := rateFilter{}
	rates := []model.ExchangeRate{
		{Time: time.Now().Add(-2 * time.Second)},
		{Time: time.Now()},
	}
	r.Check(rates)
	rates2 := []model.ExchangeRate{
		{Time: time.Now()},
		{Time: time.Now()},
	}
	r.Check(rates2)
	fmt.Println(r.Last())
}
