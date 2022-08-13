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

type service struct {
	pollPeriod            time.Duration
	currencyPairAnalyzers map[string]*CurrencyPairAnalyzer
	ohlcChannel           <-chan repo.OHLC
	generator             gs.GeneratorService
	logger                logger.Logger
	repo                  repo.Repo
}

func NewService(pollPeriod time.Duration, currencyPairs []string, timeFrames []time.Duration, generator gs.GeneratorService, logger logger.Logger, r repo.Repo) *service {

	ohlcChannel := make(chan repo.OHLC, 1)

	m := map[string]*CurrencyPairAnalyzer{}
	for _, currencyPair := range currencyPairs {
		m[currencyPair] = NewCurrencyPairAnalyzer(currencyPair, timeFrames, ohlcChannel, logger)
	}

	return &service{
		pollPeriod:            pollPeriod,
		currencyPairAnalyzers: m,
		ohlcChannel:           ohlcChannel,
		generator:             generator,
		logger:                logger,
		repo:                  r,
	}
}

func (s *service) Start(
	ctx context.Context,
	h hs.HistoryService,
	resetPeriod time.Duration) {

	//call history
	go func() {
		err := s.CollectHistory(ctx, h)
		if err != nil {
			s.logger.Error("service.Start: CollectHistory returned err: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("service.Start:  context is done; %v", ctx.Err())
			return
		default:
			s.logger.Debug("")
			s.logger.Debug("service.Start: new cycle")
			resetCtx, cancel := context.WithTimeout(ctx, resetPeriod)

			wg := sync.WaitGroup{}
			// start collecting
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.StartAnalyzers(resetCtx)
			}()

			// start polling
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := s.CollectNewRates(resetCtx)
				if err != nil {
					s.logger.Error("service.Start: CollectNewRates returned err: %v", err)
				}
			}()

			// collect OHLC and send to repo
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.CollectOHLC(resetCtx)
			}()

			wg.Wait()
			s.logger.Info("service.Start: resetCtx is done")
			s.repo.Reset(context.TODO())

			cancel()
		}
	}

}

// TODO: remove error or return meaningful errors (currently always returns nil)
func (s *service) CollectNewRates(ctx context.Context) error {
	s.logger.Debug("service.CollectNewRates: start")
	defer s.logger.Debug("service.CollectNewRates: end")

	pollTicker := time.NewTicker(s.pollPeriod)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("service.CollectNewRates: context is done: %v", ctx.Err())
			return nil
		case <-pollTicker.C:
			// TODO : add WaitGroup
			for currencyPair, analyzer := range s.currencyPairAnalyzers {

				values, err := s.generator.Values(ctx, currencyPair)
				if err != nil {
					s.logger.Error("GeneratorService.Values: %v", err)
					return err
				}
				rates := make([]ExchangeRate, len(values))
				for i := 0; i < len(rates); i++ {
					rates[i] = ExchangeRate{
						Time: values[i].Time,
						Rate: values[i].Rate,
					}
				}

				analyzer.Send(ctx, rates)
			}
		}
	}
}

var ErrResponseParse = errors.New("HistoryService responded with unparsed body")

func (s *service) CollectHistory(ctx context.Context, h hs.HistoryService) error {
	s.logger.Debug("service.CollectHistory: start")
	defer s.logger.Debug("service.CollectHistory: end")

	dayStart := time.Now().Truncate(24 * time.Hour)
	now := time.Now()

	for currencyPair, currencyPairAnalyzer := range s.currencyPairAnalyzers {
		resp, err := h.GetRatesCurrencyPairWithResponse(ctx, currencyPair, &hs.GetRatesCurrencyPairParams{
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

		currencyPairAnalyzer.Send(ctx, rates)
	}

	return nil
}

func (s *service) CollectOHLC(ctx context.Context) {
	s.logger.Debug("service.CollectOHLC: start")
	defer s.logger.Debug("service.CollectOHLC: end")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("service.CollectOHLC: context is done: %v", ctx.Err())
			return
		case ohlc, ok := <-s.ohlcChannel:
			if !ok {
				s.logger.Info("service.CollectOHLC: channel is closed")
				return
			}
			if err := s.repo.Put(ctx, ohlc.CurrencyPair, ohlc.TimeFrame, ohlc); err != nil {
				s.logger.Error("service.CollectOHLC: repo.Put err : %v", err)
			}
		}
	}
}

func (s *service) StartAnalyzers(ctx context.Context) {
	for _, analyzer := range s.currencyPairAnalyzers {
		analyzer.StartReceive(ctx)
	}
}
