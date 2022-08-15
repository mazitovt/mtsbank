// TODO: StartCollectNewRates and CollectHistory are simlilar.

package internal

import (
	"context"
	"github.com/mazitovt/logger"
	"mtsbank/analysis/internal/analyzer"
	gs "mtsbank/analysis/internal/client/generator_service"
	hs "mtsbank/analysis/internal/client/history_service"
	"mtsbank/analysis/internal/model"
	"mtsbank/analysis/internal/repo"
	"sync"
	"time"
)

type service struct {
	currencyPairAnalyzers []analyzer.Analyzer

	batchPeriod time.Duration
	batchSize   int
	resetPeriod time.Duration
	pollPeriod  time.Duration

	history   hs.HistoryService
	generator gs.GeneratorService
	repo      repo.Repo

	logger logger.Logger
}

func NewService(
	currencyPairAnalyzers []analyzer.Analyzer,
	batchPeriod time.Duration,
	batchSize int,
	resetPeriod time.Duration,
	pollPeriod time.Duration,
	history hs.HistoryService,
	generator gs.GeneratorService,
	repo repo.Repo,
	logger logger.Logger) *service {
	return &service{
		currencyPairAnalyzers: currencyPairAnalyzers,
		batchPeriod:           batchPeriod,
		batchSize:             batchSize,
		resetPeriod:           resetPeriod,
		pollPeriod:            pollPeriod,
		history:               history,
		generator:             generator,
		repo:                  repo,
		logger:                logger,
	}
}

func (s *service) Start(ctx context.Context) {
	s.logger.Debug("service.Start2: start")
	defer s.logger.Debug("service.Start2: end")

	historyCollected := make([]chan struct{}, len(s.currencyPairAnalyzers))
	for i := range historyCollected {
		historyCollected[i] = make(chan struct{}, 0)
		go func(i int) {
			historyCollected[i] <- struct{}{}
		}(i)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.logger.Info("START NEW CYCLE")

			// create new context with timeout
			resetCtx, cancel := context.WithTimeout(ctx, s.resetPeriod)

			wg := sync.WaitGroup{}
			wg.Add(len(s.currencyPairAnalyzers))
			for i := range s.currencyPairAnalyzers {
				go func(i int) {
					defer wg.Done()
					s.newCycle2(resetCtx, historyCollected[i], s.currencyPairAnalyzers[i], s.history)
				}(i)
			}
			wg.Wait()
			s.repo.Reset(context.TODO())
			cancel()
			s.logger.Info("CYCLE HAS ENDED")
		}
	}
}

func (s *service) newCycle2(ctx context.Context, historyCollected <-chan struct{}, a analyzer.Analyzer, h hs.HistoryService) {

	in := make(chan []model.ExchangeRate, 1)

	// close in channel to stop Analyzer
	out := a.Start(in)

	select {
	default:
	case <-historyCollected:
		historyCollected = nil
		s.collectHistory(ctx, in, a.CurrencyPair(), h)
	}

	// collect new rates
	go func() {
		ticker := time.NewTicker(s.pollPeriod)
		defer ticker.Stop()
		for {
			s.collectNewRates(ctx, in, a.CurrencyPair())
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				continue
			}
		}
	}()

	go func() {
		<-ctx.Done()
		close(in)
	}()

	// store to repo
	s.storeOHLC(out, s.batchPeriod, s.batchSize)
}

func (s *service) collectHistory(ctx context.Context, in chan<- []model.ExchangeRate, currencyPair string, h hs.HistoryService) {
	s.logger.Debug("service.collectHistory[%v]: start", currencyPair)
	defer s.logger.Debug("service.collectHistory[%v]: end", currencyPair)

	dayStart := time.Now().Truncate(24 * time.Hour)
	now := time.Now()

	resp, err := h.GetRatesCurrencyPairWithResponse(ctx, currencyPair, &hs.GetRatesCurrencyPairParams{
		From: &dayStart,
		To:   &now,
	})

	if err == context.Canceled || err == context.DeadlineExceeded {
		return
	} else if err != nil {
		s.logger.Error("HistoryService.GetRatesCurrencyPairWithResponse: %v", err)
		return
	}

	if resp.JSON200 == nil {
		s.logger.Warn("Couldn't extract values from response, status: %v, code: %v", resp.Status(), resp.StatusCode())
		return
	}

	s.logger.Debug("Values from HistoryService: %v", *resp.JSON200)

	rates := make([]model.ExchangeRate, len(*resp.JSON200))
	for i := range rates {
		rates[i] = model.ExchangeRate{
			Time: (*resp.JSON200)[i].Time,
			Rate: (*resp.JSON200)[i].Rate,
		}
	}
	in <- rates
}

func (s *service) collectNewRates(ctx context.Context, in chan<- []model.ExchangeRate, currencyPair string) {
	s.logger.Debug("service.collectNewRates2: start")
	defer s.logger.Debug("service.collectNewRates2: end")

	values, err := s.generator.Values(ctx, currencyPair)

	if err == context.Canceled || err == context.DeadlineExceeded {
		return
	} else if err != nil {
		s.logger.Error("GeneratorService.Values err: %v", err)
		return
	}

	rates := make([]model.ExchangeRate, len(values))
	for i := 0; i < len(rates); i++ {
		rates[i] = model.ExchangeRate{
			Time: values[i].Time,
			Rate: values[i].Rate,
		}
	}

	in <- rates
}

func (s *service) storeOHLC(in <-chan model.OHLC, batchPeriod time.Duration, batchSize int) {
	s.logger.Debug("storeOHLC2: start")
	defer s.logger.Debug("storeOHLC2: end")

	ticker := time.NewTicker(batchPeriod)
	defer ticker.Stop()

	batch := make([]model.OHLC, 0, batchSize)

	putBatch := func() {
		if err := s.repo.PutMany(context.TODO(), batch); err != nil {
			s.logger.Error("Repo.PutMany err: %v", err)
		}
	}

	for {
		select {
		case <-ticker.C:
			s.logger.Debug("storeOHLC2:: batch period is elapsed")
			if len(batch) != 0 {
				putBatch()
				batch = make([]model.OHLC, 0, batchSize)
			}
		case ohlc, ok := <-in:
			if !ok {
				s.logger.Debug("storeOHLC2:: in is closed")
				if len(batch) != 0 {
					putBatch()
				}
				return
			}
			s.logger.Info("serive.storeOHLC2: ohlc: %v", ohlc)

			batch = append(batch, ohlc)
			if len(batch) == batchSize {
				putBatch()
				batch = make([]model.OHLC, 0, batchSize)
			}
		}
	}
}
