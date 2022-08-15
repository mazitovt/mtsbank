package analyzer

import (
	"fmt"
	"github.com/mazitovt/logger"
	"mtsbank/analysis/internal/model"
	"sync"
	"time"
)

type Analyzer interface {
	CurrencyPair() string
	Start(in <-chan []model.ExchangeRate) <-chan model.OHLC
}

type TimeFrameChannel struct {
	timeFrame time.Duration
	ch        chan model.ExchangeRate
}

type CurrencyPairAnalyzer struct {
	timeFrames   []time.Duration
	currencyPair string
	logger       logger.Logger
	rateFilter   rateFilter
}

// NewCurrencyPairAnalyzer creates new CurrencyPairAnalyzer
func NewCurrencyPairAnalyzer(currencyPair string, timeFrames []time.Duration, logger logger.Logger) *CurrencyPairAnalyzer {
	return &CurrencyPairAnalyzer{
		timeFrames:   timeFrames,
		currencyPair: currencyPair,
		logger:       logger,
	}
}

// CurrencyPair returns currency pair
func (c *CurrencyPairAnalyzer) CurrencyPair() string {
	return c.currencyPair
}

func (c *CurrencyPairAnalyzer) Start(in <-chan []model.ExchangeRate) <-chan model.OHLC {
	c.logger.Debug("CurrencyPairAnalyzer.Start2: start")
	defer c.logger.Debug("CurrencyPairAnalyzer. Start2: end")

	ratesShuffleChsIn := make([]<-chan []model.ExchangeRate, len(c.timeFrames))
	ratesShuffleChsOut := make([]chan<- []model.ExchangeRate, len(c.timeFrames))
	for i := range ratesShuffleChsIn {
		ch := make(chan []model.ExchangeRate, 1)
		ratesShuffleChsIn[i] = ch
		ratesShuffleChsOut[i] = ch
	}

	go func() {
		// Receive ends when ratesCh is closed
		c.receive(in, ratesShuffleChsOut)
		for i := range ratesShuffleChsOut {
			close(ratesShuffleChsOut[i])
		}
	}()

	out := make(chan model.OHLC, 1)
	go func() {
		// Analyze create G for each ratesShuffleCh
		// Analyze ends when all Gs end
		wg := sync.WaitGroup{}
		wg.Add(len(ratesShuffleChsIn))
		for i := range ratesShuffleChsIn {
			go func(i int) {
				defer wg.Done()
				c.analyze(ratesShuffleChsIn[i], out, c.timeFrames[i])
			}(i)
		}
		wg.Wait()
		close(out)
	}()

	return out
}

func (c *CurrencyPairAnalyzer) receive(in <-chan []model.ExchangeRate, outs []chan<- []model.ExchangeRate) {
	c.logger.Debug("CurrencyPairAnalyzer.receive2: start")
	defer c.logger.Debug("CurrencyPairAnalyzer.receive2: end")
	for {
		rates, ok := <-in
		if !ok {
			return
		}
		index := c.rateFilter.Check(rates)
		rates = rates[index:]
		c.logger.Debug("CurrencyPairAnalyzer.Put: filter rates: last: %v , take from %v", c.rateFilter.Last(), index)

		c.logger.Info("CurrencyPairAnalyzer.receive2: rates: %v", rates)

		wg := sync.WaitGroup{}
		wg.Add(len(outs))
		for i := range outs {
			go func(i int) {
				defer wg.Done()
				outs[i] <- rates
			}(i)
		}
		wg.Wait()
	}
}

func (c *CurrencyPairAnalyzer) analyze(in <-chan []model.ExchangeRate, out chan<- model.OHLC, timeFrame time.Duration) {
	log := func(f func(format string, v ...any), format string, v ...any) {
		f(fmt.Sprintf("CurrencyPairAnalyzer.analyze2[%s]: ", timeFrame.String())+format, v...)
	}
	log(c.logger.Debug, "%s", "start")
	defer log(c.logger.Debug, "%s", "end")

	ohlc := model.NewOHLC(c.currencyPair, timeFrame)

	for {
		rates, ok := <-in
		if !ok {
			if !ohlc.Empty() {
				log(c.logger.Debug, "%v", *ohlc)
				out <- *ohlc
			}
			return
		}
		log(c.logger.Info, "rates: %v", rates)

		for i := range rates {
			if ohlc.UpdateOrReady(rates[i]) {
				log(c.logger.Debug, "%v", *ohlc)
				out <- *ohlc
				ohlc.Reset()
				ohlc.UpdateOrReady(rates[i])
			}
		}
	}
}
