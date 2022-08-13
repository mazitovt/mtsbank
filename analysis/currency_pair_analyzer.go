package analysis

import (
	"context"
	"fmt"
	"mtsbank/analysis/internal/repo"
	"mtsbank/analysis/logger"
	"sync"
	"time"
)

type rateFilter struct {
	last time.Time
}

// Check takes ORDERED rates and return index, from which starts unseen data
func (t *rateFilter) Check(rates []ExchangeRate) int64 {
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

type TimeFrameChannel struct {
	timeFrame time.Duration
	ch        chan ExchangeRate
}

type CurrencyPairAnalyzer struct {
	currencyPair      string
	logger            logger.Logger
	timeFrameChannels []TimeFrameChannel
	out               chan<- repo.OHLC
	rateFilter        rateFilter
}

func NewCurrencyPairAnalyzer(currencyPair string, timeFrames []time.Duration, out chan<- repo.OHLC, logger logger.Logger) *CurrencyPairAnalyzer {
	timeFramesChannels := make([]TimeFrameChannel, len(timeFrames))
	for i, timeFrame := range timeFrames {
		timeFramesChannels[i] = TimeFrameChannel{
			timeFrame: timeFrame,
			ch:        make(chan ExchangeRate, 1),
		}
	}
	return &CurrencyPairAnalyzer{
		currencyPair:      currencyPair,
		logger:            logger,
		timeFrameChannels: timeFramesChannels,
		out:               out,
	}
}

func (c *CurrencyPairAnalyzer) closeChannels() {
	for _, t := range c.timeFrameChannels {
		close(t.ch)
	}
}

// Send reads rates and send them to timeFrameChannels
//
// Blocked until rates sent to every channel
func (c *CurrencyPairAnalyzer) Send(ctx context.Context, rates []ExchangeRate) {
	c.logger.Debug("CurrencyPairAnalyzer.Send: start")
	defer c.logger.Debug("CurrencyPairAnalyzer.Send: end")

	c.logger.Debug("rates before: %v", rates)

	index := c.rateFilter.Check(rates)
	rates = rates[index:]
	c.logger.Info("Filter rates: last: %v , take from %v", c.rateFilter.Last(), index)

	c.logger.Debug("rates after: %v", rates)

	wg := sync.WaitGroup{}
	wg.Add(len(c.timeFrameChannels))

	select {
	case <-ctx.Done():
		c.logger.Info("CurrencyPairAnalyzer.Send: no values were sent to channel, context.Err(): %v", ctx.Err())
		c.logger.Debug("Close channels, currencyPair: %s", c.currencyPair)
		c.closeChannels()
		return
	default:
		for _, timeFrameChannel := range c.timeFrameChannels {

			// G reads rates and sends to ch
			go func(ch chan<- ExchangeRate) {
				defer wg.Done()
				for i := 0; i < len(rates); i++ {
					ch <- rates[i]
				}
			}(timeFrameChannel.ch)
		}
	}

	c.logger.Debug("CurrencyPairAnalyzer.Send: waiting on WaitGroup")
	wg.Wait()
	c.logger.Debug("CurrencyPairAnalyzer.Send: done waiting on WaitGroup")
}

// StartReceive starts all receive goroutines
func (c *CurrencyPairAnalyzer) StartReceive(ctx context.Context) {
	for _, timeFrameChannel := range c.timeFrameChannels {
		go c.receive(ctx, timeFrameChannel.ch, timeFrameChannel.timeFrame)
	}
}

func (c *CurrencyPairAnalyzer) receive(ctx context.Context, in <-chan ExchangeRate, period time.Duration) {
	loggerLine := fmt.Sprintf("CurrencyPairAnalyzer.receive[%s]", period.String())
	log := func(f func(format string, v ...any), format string, v ...any) {
		f("%s: "+format, loggerLine, v)
	}

	log(c.logger.Debug, "%s", "start")
	defer log(c.logger.Debug, "%s", "end")

	calculatorOHLC := NewCalculatorOHLC(c.currencyPair, period.String())

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {

		// Additional select to prioritize context and ticker
		select {
		case <-ctx.Done():
			log(c.logger.Info, "Context is done, err: %v", ctx.Err())
			return
		case <-ticker.C:
			log(c.logger.Info, "Collect OHLC")
			if calculatorOHLC.IsEmpty() {
				log(c.logger.Info, "No values were generated")
				continue
			}
			c.out <- calculatorOHLC.OHLC()
			calculatorOHLC.Reset()
		default:
		}

		select {
		case <-ctx.Done():
			log(c.logger.Info, "Context is done, err: %v", ctx.Err())
			return
		case <-ticker.C:
			log(c.logger.Info, "Collect OHLC")
			if calculatorOHLC.IsEmpty() {
				log(c.logger.Info, "No values were generated")
				continue
			}
			c.out <- calculatorOHLC.OHLC()
			calculatorOHLC.Reset()
		case r, ok := <-in:
			if !ok {
				log(c.logger.Info, "Data channel is closed")
				return
			}
			log(c.logger.Debug, "Data received: %v", r)
			calculatorOHLC.Update(r)
		}
	}
}
