package internal

import (
	"fmt"
	"testing"
	"time"
)

func TestPriceGenerator(t *testing.T) {

	pairs := []string{"EURUSD", "USDRUB", "USDJPY"}
	period := 1000 * time.Millisecond
	g := NewSimplePriceGenerator(pairs, ExchangeRateFromTime, 3)
	go func() {
		for {
			g.Generate()
			time.Sleep(period)
		}
	}()

	time.Sleep(period / 2)

	for i := 0; i < 10; i++ {
		v, _ := g.Get(pairs[0])
		fmt.Println(v)
		v, _ = g.Get(pairs[1])
		fmt.Println(v)
		v, _ = g.Get(pairs[2])
		fmt.Println(v)
		_ = v
		time.Sleep(period)
	}

}
