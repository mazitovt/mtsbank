// TODO: StartCollectNewRates and CollectHistory are simlilar.
// TODO: ticker isn't suitable bc it wait for time before first tick

package analysis

import (
	"context"
	"errors"
	analyzer "mtsbank/analysis/internal/analyzer"
	gs "mtsbank/analysis/internal/client/generator_service"
	hs "mtsbank/analysis/internal/client/history_service"
	"mtsbank/analysis/internal/model"
	"mtsbank/analysis/internal/repo"
	"mtsbank/analysis/logger"
	"sync"
	"time"
)

type service struct {
	pollPeriod            time.Duration
	currencyPairAnalyzers []analyzer.Analyzer
	generator             gs.GeneratorService
	logger                logger.Logger
	repo                  repo.Repo
}

func NewService(pollPeriod time.Duration, currencyPairs []string, timeFrames []time.Duration, generator gs.GeneratorService, logger logger.Logger, r repo.Repo) *service {

	s := make([]analyzer.Analyzer, len(currencyPairs))
	for i, currencyPair := range currencyPairs {
		s[i] = analyzer.NewCurrencyPairAnalyzer(currencyPair, timeFrames, r, logger)
	}

	return &service{
		pollPeriod:            pollPeriod,
		currencyPairAnalyzers: s,
		generator:             generator,
		logger:                logger,
		repo:                  r,
	}
}

func (s *service) Start(ctx context.Context, h hs.HistoryService, resetPeriod time.Duration, batchPeriod time.Duration, batchSize int) {
	s.logger.Debug("service.Start: start")
	defer s.logger.Debug("service.Start: end")

	historyCollected := make(chan struct{}, 0)
	go func() {
		historyCollected <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// create new context with timeout
			resetCtx, cancel := context.WithTimeout(ctx, resetPeriod)

			s.logger.Info("START NEW CYCLE")

			wg := sync.WaitGroup{}
			wg.Add(len(s.currencyPairAnalyzers))
			for i := range s.currencyPairAnalyzers {
				go func(i int) {
					defer wg.Done()
					s.currencyPairAnalyzers[i].Start(resetCtx, batchPeriod, batchSize)
				}(i)
			}

			select {
			case <-historyCollected:
				historyCollected = nil
				s.CollectHistory(resetCtx, h)
			default:
			}

			go s.StartCollectNewRates(resetCtx)

			wg.Wait()

			s.repo.Reset(context.TODO())

			cancel()

			s.logger.Info("CYCLE IS OVER")
		}
	}
}

var ErrResponseParse = errors.New("HistoryService responded with unparsed body")

// StartCollectNewRates periodically request new rates from generator and sends to analyzers
// StartCollectNewRates is not thread-safe
// Blocks until context is done
func (s *service) StartCollectNewRates(ctx context.Context) {
	s.logger.Info("service.StartCollectNewRates: start")
	defer s.logger.Info("service.StartCollectNewRates: end")

	pollTicker := time.NewTicker(s.pollPeriod)
	defer pollTicker.Stop()

	select {
	case <-ctx.Done():
		return
	default:

	}

	for {

		wg := sync.WaitGroup{}
		wg.Add(len(s.currencyPairAnalyzers))

		for index := range s.currencyPairAnalyzers {
			go func(index int) {
				defer wg.Done()
				a := s.currencyPairAnalyzers[index]
				values, err := s.generator.Values(ctx, a.CurrencyPair())
				if err != nil {
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

				a.Put(ctx, rates)
			}(index)
		}

		wg.Wait()
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			continue
		}
	}
}

// CollectHistory request all rates from begging of this day to current time
// CollectHistory is not thread-safe
// Blocks until all analysers received their all values
func (s *service) CollectHistory(ctx context.Context, h hs.HistoryService) {
	s.logger.Info("service.CollectHistory: start")
	defer s.logger.Info("service.CollectHistory: end")

	dayStart := time.Now().Truncate(24 * time.Hour)
	now := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(len(s.currencyPairAnalyzers))

	for index := range s.currencyPairAnalyzers {
		go func(index int) {
			defer wg.Done()
			a := s.currencyPairAnalyzers[index]
			resp, err := h.GetRatesCurrencyPairWithResponse(ctx, a.CurrencyPair(), &hs.GetRatesCurrencyPairParams{
				From: &dayStart,
				To:   &now,
			})
			if err != nil {
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

			s.currencyPairAnalyzers[index].Put(ctx, rates)
		}(index)
	}

	wg.Wait()
}
