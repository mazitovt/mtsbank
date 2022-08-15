package internal

import (
	"context"
	"encoding/json"
	"github.com/mazitovt/logger"
	"golang.org/x/sync/errgroup"
	api "mtsbank/history/internal/api/http/v1"
	gs "mtsbank/history/internal/client/generator_service"
	"mtsbank/history/internal/repo"
	"net/http"
	"sync"
	"time"
)

var poolExchangeRates = sync.Pool{New: func() any {
	return []api.ExchangeRate{}
}}

var _ api.ServerInterface = (*SimpleHistoryService)(nil)

type SimpleHistoryService struct {
	repo            repo.Repo
	generatorClient gs.GeneratorService
	logger          logger.Logger
}

func NewSimpleHistoryService(repo repo.Repo, generatorClient gs.GeneratorService, logger logger.Logger) *SimpleHistoryService {
	return &SimpleHistoryService{repo: repo, generatorClient: generatorClient, logger: logger}
}

func (s *SimpleHistoryService) GetRatesCurrencyPair(w http.ResponseWriter, r *http.Request, currencyPair string, params api.GetRatesCurrencyPairParams) {

	if params.From == nil || params.To == nil {
		s.writeError(w, http.StatusBadRequest, "incorrect time format")
		return
	}

	res, err := s.repo.GetByTime(r.Context(), currencyPair, *params.From, *params.To)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	exchangeRates := make([]api.ExchangeRate, len(res))
	for i := range res {
		exchangeRates[i] = api.ExchangeRate{
			Time: res[i].Time,
			Rate: res[i].Rate,
		}
	}

	// content is set to application/json only that order
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(exchangeRates); err != nil {
		s.logger.Error("SimpleHistoryService.writeError: err: %v", err)
	}
}

func (s *SimpleHistoryService) Start(ctx context.Context, period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		currencies, err := s.repo.Currencies(ctx)
		if err != nil {
			s.logger.Error("repo.Repo.Currencies err: %v", err)
		}

		g := new(errgroup.Group)

		for _, c := range currencies {
			g.Go(func() error {
				var err error
				out := poolExchangeRates.Get().([]api.ExchangeRate)
				out = out[:0]
				defer poolExchangeRates.Put(out)
				out, err = s.generatorClient.GetRates(ctx, c, out)
				if err != nil && err != gs.ErrBufferGrow {
					return err
				}
				return s.repo.InsertWithCurrencyPair(ctx, c, out)
			})
		}

		if err = g.Wait(); err != nil {
			s.logger.Error("collecting new rates : err: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			continue
		}
	}
}

func (s *SimpleHistoryService) writeError(w http.ResponseWriter, code int, message string) {
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

func (s *SimpleHistoryService) getByTime(ctx context.Context, currencyPair string, start time.Time, end time.Time) ([]api.ExchangeRate, error) {
	res, err := s.repo.GetByTime(ctx, currencyPair, start, end)
	if err != nil {
		s.logger.Error("Repo.GetByTime: %v", err)
		return nil, err
	}
	exchangeRates := make([]api.ExchangeRate, 0)
	for _, rate := range res {
		exchangeRates = append(exchangeRates, api.ExchangeRate{
			Time: rate.Time,
			Rate: rate.Rate,
		})
	}

	return exchangeRates, nil
}
