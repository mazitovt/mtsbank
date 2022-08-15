package internal

import (
	"context"
	"github.com/mazitovt/logger"
	"testing"
	"time"
)

func TestPriceGenerator(t *testing.T) {

	pairs := []string{"EURUSD", "USDRUB", "USDJPY"}
	period := 1000 * time.Millisecond
	g := NewSimplePriceGenerator(pairs, ExchangeRateFromTime, 3, logger.New(logger.Info))

	ctx, cancel := context.WithCancel(context.Background())

	go g.Start(ctx, period)

	time.Sleep(3 * time.Second)
	cancel()

	time.Sleep(3 * time.Second)

}
