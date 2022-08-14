package analyzer

import (
	"context"
	"fmt"
	"mtsbank/analysis/internal/model"
	"mtsbank/analysis/internal/repo"
	"mtsbank/analysis/logger"
	"sync"
	"time"
)

type rateFilter struct {
	last time.Time
}

type Analyzer interface {
	Put(ctx context.Context, rates []model.ExchangeRate)
	Start(ctx context.Context, batchPeriod time.Duration, batchSize int)
	CurrencyPair() string
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

type TimeFrameChannel struct {
	timeFrame time.Duration
	ch        chan model.ExchangeRate
}

type CurrencyPairAnalyzer struct {
	currencyPair          string
	logger                logger.Logger
	timeFrameChannels     []TimeFrameChannel
	out                   chan model.OHLC
	rateFilter            rateFilter
	repo                  repo.Repo
	writingToChannelsLock sync.Locker
}

// NewCurrencyPairAnalyzer creates new CurrencyPairAnalyzer
func NewCurrencyPairAnalyzer(currencyPair string, timeFrames []time.Duration, r repo.Repo, logger logger.Logger) *CurrencyPairAnalyzer {

	timeFrameChannels := make([]TimeFrameChannel, len(timeFrames))

	for i, timeFrame := range timeFrames {
		timeFrameChannels[i] = TimeFrameChannel{
			timeFrame: timeFrame,
		}
	}

	return &CurrencyPairAnalyzer{
		rateFilter:            rateFilter{},
		timeFrameChannels:     timeFrameChannels,
		currencyPair:          currencyPair,
		repo:                  r,
		logger:                logger,
		writingToChannelsLock: &sync.Mutex{},
	}
}

// CurrencyPair returns currency pair
func (c *CurrencyPairAnalyzer) CurrencyPair() string {
	return c.currencyPair
}

// Start initializes channels, start receiving Gs and waits for store
// Blocked until store is done
func (c *CurrencyPairAnalyzer) Start(ctx context.Context, batchPeriod time.Duration, batchSize int) {
	c.init()

	go func() {
		<-ctx.Done()
		c.writingToChannelsLock.Lock()
		defer c.writingToChannelsLock.Unlock()
		for _, t := range c.timeFrameChannels {
			close(t.ch)
		}
		c.logger.Info("close time frame channels")
	}()

	go c.startAnalyze()
	c.store(c.out, batchPeriod, batchSize)
}

// Put puts rates into pipeline.
// Blocked until rates are sent.
// Put is not thread-safe.
func (c *CurrencyPairAnalyzer) Put(ctx context.Context, rates []model.ExchangeRate) {
	log := func(f func(format string, v ...any), format string, v ...any) {
		f(fmt.Sprintf("CurrencyPairAnalyzer.Put[%s]: ", c.currencyPair)+format, v...)
	}

	log(c.logger.Debug, "%s", "start")
	defer log(c.logger.Debug, "%s", "end")

	index := c.rateFilter.Check(rates)
	rates = rates[index:]
	log(c.logger.Debug, "CurrencyPairAnalyzer.Put: filter rates: last: %v , take from %v", c.rateFilter.Last(), index)

	select {
	case <-ctx.Done():
		return
	default:
		wg := sync.WaitGroup{}
		wg.Add(len(c.timeFrameChannels))

		for _, timeFrameChannel := range c.timeFrameChannels {

			// G reads rates and sends to ch
			go func(ch chan<- model.ExchangeRate) {
				defer wg.Done()
				c.writingToChannelsLock.Lock()
				defer c.writingToChannelsLock.Unlock()
				for i := 0; i < len(rates); i++ {
					ch <- rates[i]
				}
			}(timeFrameChannel.ch)
		}
		c.logger.Debug("CurrencyPairAnalyzer.Put: waiting on WaitGroup")
		wg.Wait()
		c.logger.Debug("CurrencyPairAnalyzer.Put: done waiting on WaitGroup")
	}
}

// startAnalyze starts and waits analyze goroutines to end
func (c *CurrencyPairAnalyzer) startAnalyze() {
	wg := sync.WaitGroup{}

	for _, timeFrameChannel := range c.timeFrameChannels {
		wg.Add(1)
		timeFrameChannel := timeFrameChannel
		go func() {
			defer wg.Done()
			c.analyze(c.out, timeFrameChannel.ch, timeFrameChannel.timeFrame)
		}()
	}

	wg.Wait()
	close(c.out)
	c.logger.Debug("CurrencyPairAnalyzer.StartReceive[%v]: close out channel", c.currencyPair)
}

// Reads from in and write to out
//
// Blocks until in is closed. When in is closed, closes out
func (c *CurrencyPairAnalyzer) analyze(out chan<- model.OHLC, in <-chan model.ExchangeRate, period time.Duration) {
	log := func(f func(format string, v ...any), format string, v ...any) {
		f(fmt.Sprintf("CurrencyPairAnalyzer.analyze[%s]: ", period.String())+format, v...)
	}

	log(c.logger.Debug, "%s", "start")
	defer log(c.logger.Debug, "%s", "end")

	ohlc := model.OHLC{
		CurrencyPair: c.currencyPair,
		TimeFrame:    period,
	}

	for {
		r, ok := <-in
		if !ok {
			return
		}
		log(c.logger.Debug, "Data received: %v", r)

		if ohlc.UpdateOrReady(r) {
			log(c.logger.Info, "%s", "put ohlc")
			out <- ohlc
			ohlc.Reset()
			ohlc.UpdateOrReady(r)
			continue
		}
	}
}

// Reads from in.
// Blocks until in is closed
func (c *CurrencyPairAnalyzer) store(in <-chan model.OHLC, batchPeriod time.Duration, batchSize int) {
	log := func(f func(format string, v ...any), format string, v ...any) {
		f(fmt.Sprintf("CurrencyPairAnalyzer.store[%v]: ", c.currencyPair)+format, v...)
	}

	log(c.logger.Debug, "%s", "start")
	defer log(c.logger.Debug, "%s", "end")

	ticker := time.NewTicker(batchPeriod)
	defer ticker.Stop()

	batch := make([]model.OHLC, 0, batchSize)

	putBatch := func() {
		log(c.logger.Info, "%s", "put batch")
		if err := c.repo.PutMany(context.TODO(), batch); err != nil {
			log(c.logger.Error, "Repo.PutMany err: %v", err)
		}
	}

	for {
		select {
		case <-ticker.C:
			if len(batch) != 0 {
				putBatch()
				batch = make([]model.OHLC, 0, batchSize)
			}
		case ohlc, ok := <-in:
			log(c.logger.Debug, "Data received: %v", ohlc)
			if !ok {
				if len(batch) != 0 {
					putBatch()
				}
				return
			}

			batch = append(batch, ohlc)
			if len(batch) == batchSize {
				putBatch()
				batch = make([]model.OHLC, 0, batchSize)
			}
		}
	}
}

// init creates new out channel and new time frame channels
func (c *CurrencyPairAnalyzer) init() {
	for i := range c.timeFrameChannels {
		c.timeFrameChannels[i].ch = make(chan model.ExchangeRate, 1)
	}

	c.out = make(chan model.OHLC, 1)
}
