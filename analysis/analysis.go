package analysis

import (
	"context"
	"errors"
	gs "mtsbank/analysis/internal/client/generator_service"
	hs "mtsbank/analysis/internal/client/history_service"
	"mtsbank/analysis/internal/repo"
	"mtsbank/analysis/logger"
	"sync"
	"time"
)

type ExchangeRate struct {
	Time time.Time
	Rate int64
}

type Option interface {
	apply()
}

type Analyzer interface {
	Get(ctx context.Context, currencyPair string, timeFrame string, options ...Option)
	GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64) ([]repo.OHLC, error)

	// Start in G
	CollectNewRates(ctx context.Context)
	// Start in G
	CollectHistory(ctx context.Context, h hs.HistoryService) error
}

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

type service struct {
	rateFilter rateFilter

	currencyPairs []string
	timeFrames    []time.Duration

	//TODO: when to close channels?
	currencyPairToTimeFrameChannels map[string][]TimeFrameChannel
	generator                       gs.GeneratorService
	logger                          logger.Logger
	repo                            repo.Repo
}

type TimeFrameChannel struct {
	timeFrame time.Duration
	ch        chan ExchangeRate
}

func NewService(currencyPairs []string, timeFrames []time.Duration, generator gs.GeneratorService, logger logger.Logger, repo repo.Repo) *service {

	m := map[string][]TimeFrameChannel{}
	for _, currencyPair := range currencyPairs {
		timeFramesChannels := make([]TimeFrameChannel, len(timeFrames))
		for i, timeFrame := range timeFrames {
			timeFramesChannels[i] = TimeFrameChannel{
				timeFrame: timeFrame,
				ch:        make(chan ExchangeRate, 1),
			}
		}
		m[currencyPair] = timeFramesChannels
	}

	return &service{
		rateFilter:                      rateFilter{},
		currencyPairs:                   currencyPairs,
		timeFrames:                      timeFrames,
		currencyPairToTimeFrameChannels: m,
		generator:                       generator,
		logger:                          logger,
		repo:                            repo,
	}
}

// StartTimeFrameGoroutines starts service with reset period. BLOCKING
func (s *service) StartTimeFrameGoroutines(ctx context.Context, resetPeriod time.Duration) {

	resetCtx, cancel := context.WithTimeout(ctx, resetPeriod)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context is done: %v", ctx.Err())
			return
		default:
			wg := sync.WaitGroup{}

			goroutineCount := len(s.timeFrames) * len(s.currencyPairs)
			wg.Add(goroutineCount)

			s.logger.Info("Starting TimeFrameGoroutines")
			for currencyPair, timeFrameChannels := range s.currencyPairToTimeFrameChannels {
				for _, timeFrameChannel := range timeFrameChannels {
					go func(t TimeFrameChannel) {
						defer wg.Done()
						s.timeFrameGoroutine(resetCtx, currencyPair, t.ch, t.timeFrame)
					}(timeFrameChannel)
				}
			}

			s.logger.Info("Waiting on TimeFrameGoroutines: start")
			wg.Wait()
			s.logger.Info("Waiting on TimeFrameGoroutines: end")
		}

		s.logger.Info("All TimeFrameGoroutines are done")
	}

}

// CollectNewRates periodically ask generator for new rates. CollectNewRates is a blocking function (run in G)
func (s *service) CollectNewRates(ctx context.Context, period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context is done, err: %v", ctx.Err())
			return
		case <-ticker.C:
			s.logger.Debug("Collecting new rates from Generator: start")
			if err := s.collectRates(ctx); err != nil {
				s.logger.Error("service.collectRates: %v", err)
				//TODO: return or not?
			}
			s.logger.Info("Collecting new rates from Generator: done")
		}
	}
}

// TODO: remove error or return meaningful errors (currently always returns nil)
// TODO:
func (s *service) collectRates(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(len(s.currencyPairs))

	for _, currency := range s.currencyPairs {
		go func(currency string) {
			defer wg.Done()
			values, err := s.generator.Values(ctx, currency)
			if err != nil {
				s.logger.Error("GeneratorService.Values: %v", err)
				return
			}
			rates := make([]ExchangeRate, len(values))
			for i := 0; i < len(rates); i++ {
				rates[i] = ExchangeRate{
					Time: values[i].Time,
					Rate: values[i].Rate,
				}
			}

			err = s.sendToTimeFrameGoroutines(ctx, currency, rates)
			if err != nil {
				s.logger.Error("service.SendToTimeFrameGoroutines: %v", err)
				return
			}

		}(currency)
	}

	wg.Wait()

	return nil
}

var ErrResponseParse = errors.New("HistoryService responded with unparsed body")

func (s *service) CollectHistory(ctx context.Context, h hs.HistoryService) error {
	dayStart := time.Now().Truncate(24 * time.Hour)
	now := time.Now()
	for _, currency := range s.currencyPairs {
		resp, err := h.GetRatesCurrencyPairWithResponse(ctx, currency, &hs.GetRatesCurrencyPairParams{
			From: &dayStart,
			To:   &now,
		})
		if err != nil {
			s.logger.Error("HistoryService.GetRatesCurrencyPairWithResponse: %v", err)
			return err
		}
		if resp.JSON200 == nil {
			s.logger.Warn("Couldn't extract values from response, status: %v, code: %v", resp.Status(), resp.StatusCode())
			return ErrResponseParse
		}

		s.logger.Info("Values from HistoryService: %v", *resp.JSON200)

		rates := make([]ExchangeRate, len(*resp.JSON200))
		for i := 0; i < len(rates); i++ {
			rates[i] = ExchangeRate{
				Time: (*resp.JSON200)[i].Time,
				Rate: (*resp.JSON200)[i].Rate,
			}
		}

		err = s.sendToTimeFrameGoroutines(ctx, currency, rates)
		if err != nil {
			s.logger.Debug("service.SendToTimeFrameGoroutines: %v", err)
			return err
		}
	}

	return nil
}

// timeFrameGoroutine is a blocking function.
//
// Data channel requirements:
//	- ExchangeRate.Time is newer than previously sent value
// 	- ExchangeRate was not sent before
func (s *service) timeFrameGoroutine(ctx context.Context, currencyPair string, data <-chan ExchangeRate, period time.Duration) {

	calculatorOHLC := NewCalculatorOHLC()

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {

		// Additional select to prioritize context and ticker
		select {
		case <-ctx.Done():
			s.logger.Info("Context is done, err: %v", ctx.Err())
			return
		case <-ticker.C:
			s.logger.Info("Collect OHLC for %v", period)
			if calculatorOHLC.IsEmpty() {
				s.logger.Info("TimeFrame [%v]: no values were generated", period)
				continue
			}
			if err := s.repo.Put(ctx, currencyPair, period.String(), calculatorOHLC.OHLC()); err != nil {
				s.logger.Error("Repo.Put: %v", err)
				//TODO: return or not?
			}
			calculatorOHLC.Reset()
		default:
		}

		select {
		case <-ctx.Done():
			s.logger.Info("Context is done, err: %v", ctx.Err())
			return
		case <-ticker.C:
			s.logger.Info("Collect OHLC for %v", period)
			if calculatorOHLC.IsEmpty() {
				s.logger.Info("TimeFrame [%v]: no values were generated", period)
				continue
			}
			if err := s.repo.Put(ctx, currencyPair, period.String(), calculatorOHLC.OHLC()); err != nil {
				s.logger.Error("Repo.Put: %v", err)
				//TODO: return or not?
			}
			calculatorOHLC.Reset()
		case r, ok := <-data:
			if !ok {
				s.logger.Info("Data channel is closed for %v", period)
				return
			}
			s.logger.Debug("Data recieved: %v", r)
			calculatorOHLC.Update(r)
		}
	}
}

// TODO:rates should contain only new data and order by time (from old to new)
func (s *service) sendToTimeFrameGoroutines(ctx context.Context, currencyPair string, rates []ExchangeRate) error {
	s.logger.Debug("sendToTimeFrameGoroutines: start, currencyPair: %v, rates: %v", currencyPair, rates)
	defer s.logger.Debug("sendToTimeFrameGoroutines: end")

	index := s.rateFilter.Check(rates)
	rates = rates[index:]
	s.logger.Info("Filter rates: take from %v", index)

	select {
	case <-ctx.Done():
		s.logger.Info("Context is done: %v", ctx.Err())
		return ctx.Err()
	default:
		for _, timeFrameChannels := range s.currencyPairToTimeFrameChannels[currencyPair] {
			for _, rate := range rates {
				timeFrameChannels.ch <- rate
			}
		}
	}

	return nil
}

func (s *service) Get(ctx context.Context, currencyPair string, timeFrame string, options ...Option) {
	//TODO implement me
	panic("implement me")
}

func (s *service) GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64) ([]repo.OHLC, error) {
	//TODO implement me
	panic("implement me")
}

func writeCurrencyRates(ctx context.Context, g gs.GeneratorService, name string, writeChans []chan<- ExchangeRate) {

}

func readCurrencyRates(ctx context.Context, readChans <-chan ExchangeRate) {

}
