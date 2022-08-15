// TODO: StartCollectNewRates and CollectHistory are similar.

package internal

import (
	"context"
	"encoding/json"
	"github.com/mazitovt/logger"
	"github.com/sosodev/duration"
	"mtsbank/analysis/internal/analyzer"
	api "mtsbank/analysis/internal/api/http/v1"
	gs "mtsbank/analysis/internal/client/generator_service"
	hs "mtsbank/analysis/internal/client/history_service"
	"mtsbank/analysis/internal/model"
	"mtsbank/analysis/internal/repo"
	"net/http"
	"sync"
	"time"
)

var poolOHLC = sync.Pool{New: func() any {
	return []model.OHLC{}
}}

var poolOHLCDTO = sync.Pool{New: func() any {
	return []api.OHLC{}
}}

var _ api.ServerInterface = (*service)(nil)

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

func (s *service) GetRatesCurrencyPairTimeFrame(w http.ResponseWriter, r *http.Request, currencyPair string, timeFrame string, params api.GetRatesCurrencyPairTimeFrameParams) {

	d, err := duration.Parse(timeFrame)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid time frame (check ISO 8601)")
		return
	}

	timeFrame = d.ToTimeDuration().String()

	buffer := poolOHLC.Get().([]model.OHLC)
	buffer = buffer[:0]
	defer poolOHLC.Put(buffer)

	out := poolOHLCDTO.Get().([]api.OHLC)
	out = out[:0]
	defer poolOHLCDTO.Put(out)

	writeRepoErr := func(err error) {
		switch err {
		case repo.ErrBadTimeFrame, repo.ErrBadCurrencyPair, repo.ErrBadCurrencyPairOrTimeFrame:
			s.writeError(w, http.StatusBadRequest, err.Error())
		default:
			s.writeError(w, http.StatusInternalServerError, "internal error")
		}
	}

	//TODO: define rules in swagger
	switch {
	case params.Last != nil:
		buffer, err = s.repo.GetMany(r.Context(), currencyPair, timeFrame, *params.Last, buffer)
		if err != nil {
			writeRepoErr(err)
			return
		}
		for i := range buffer {
			out = append(out, api.OHLC{
				Close:     buffer[i].Close,
				CloseTime: buffer[i].CloseTime,
				High:      buffer[i].High,
				Low:       buffer[i].Low,
				Open:      buffer[i].Open,
				OpenTime:  buffer[i].OpenTime,
			})
		}
	case params.To != nil && params.From != nil:
		buffer, err = s.repo.GetMany(r.Context(), currencyPair, timeFrame, *params.Last, buffer)
		if err != nil {
			writeRepoErr(err)
			return
		}
		for i := range buffer {
			out = append(out, api.OHLC{
				Close:     buffer[i].Close,
				CloseTime: buffer[i].CloseTime,
				High:      buffer[i].High,
				Low:       buffer[i].Low,
				Open:      buffer[i].Open,
				OpenTime:  buffer[i].OpenTime,
			})
		}
	default:
		ohlc, err := s.repo.GetLast(r.Context(), currencyPair, timeFrame)
		if err != nil {
			writeRepoErr(err)
			return
		}

		out = append(out, api.OHLC{
			Close:     ohlc.Close,
			CloseTime: ohlc.CloseTime,
			High:      ohlc.High,
			Low:       ohlc.Low,
			Open:      ohlc.Open,
			OpenTime:  ohlc.OpenTime,
		})
	}

	// content is set to application/json only that order
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(out); err != nil {
		s.logger.Error("SimpleHistoryService.writeError: err: %v", err)
	}
}

func (s *service) writeError(w http.ResponseWriter, code int, message string) {
	petErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(petErr)
	if err != nil {
		s.logger.Error("SimpleHistoryService.writeError: err: %v", err)
	}
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

	//TODO: make use of sync.Pool
	//
	//cant use sync.Pool bc writing slice to channel in (reading side should put slice back to pool somehow)
	out := make([]model.ExchangeRate, 0)
	out, err := s.generator.GetRates(ctx, currencyPair, out)
	if err != nil && err != gs.ErrBufferGrow {
		s.logger.Error("cant collect new rates: %v", err)
		return
	}

	in <- out
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
