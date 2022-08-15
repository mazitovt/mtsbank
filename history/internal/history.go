package internal

import (
	"context"
	"encoding/json"
	"github.com/mazitovt/logger"
	v1 "mtsbank/history/internal/api/http/v1"
	gs "mtsbank/history/internal/client/generator_service"
	"mtsbank/history/internal/repo"
	"net/http"
	"time"
)

type ExchangeRate struct {
	Time time.Time `json:"time"`
	Rate int64     `json:"rate"`
}

type HistoryService interface {
	GetByTime(ctx context.Context, currencyPair string, start time.Time, end time.Time) ([]ExchangeRate, error)
	CollectExchangeRates(ctx context.Context) error
}

type SimpleHistoryService struct {
	repo            repo.Repo
	generatorClient gs.GeneratorService
	logger          logger.Logger
}

func (s *SimpleHistoryService) GetRatesCurrencyPair(w http.ResponseWriter, r *http.Request, currencyPair string, params v1.GetRatesCurrencyPairParams) {

	if params.From == nil || params.To == nil {
		writeError(w, http.StatusBadRequest, "incorrect time format")
		return
	}

	s.logger.Debug("%v", params)
	//TODO: check if currency_pair exists in db
	res, err := s.repo.GetByTime(r.Context(), currencyPair, *params.From, *params.To)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	exchangeRates := make([]v1.ExchangeRate, 0)
	for _, rate := range res {
		exchangeRates = append(exchangeRates, v1.ExchangeRate{
			Time: rate.Time,
			Rate: rate.Rate,
		})
	}

	// content is set to application/json only that order
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(exchangeRates)
	s.logger.Debug("Encode.Err: %v", err)
}

func writeError(w http.ResponseWriter, code int, message string) {
	petErr := v1.Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(petErr)
}

func NewSimpleHistoryService(repo repo.Repo, generatorClient gs.GeneratorService, logger logger.Logger) *SimpleHistoryService {
	return &SimpleHistoryService{repo: repo, generatorClient: generatorClient, logger: logger}
}

func (s *SimpleHistoryService) GetByTime(ctx context.Context, currencyPair string, start time.Time, end time.Time) ([]ExchangeRate, error) {
	res, err := s.repo.GetByTime(ctx, currencyPair, start, end)
	if err != nil {
		s.logger.Error("Repo.GetByTime: %v", err)
		return nil, err
	}
	exchangeRates := make([]ExchangeRate, 0)
	for _, rate := range res {
		exchangeRates = append(exchangeRates, ExchangeRate{
			Time: rate.Time,
			Rate: rate.Rate,
		})
	}

	return exchangeRates, nil
}

func (s *SimpleHistoryService) CollectExchangeRates(ctx context.Context) error {
	currencies, err := s.repo.Currencies(ctx)

	if err != nil {
		s.logger.Error("Repo.Currencies: %v", err)
		return err
	}

	// TODO: grow slice first
	rates := make([]repo.RegistryRow, 0)

	for _, c := range currencies {
		values, err := s.generatorClient.Values(ctx, c)
		if err != nil {
			s.logger.Error("GeneratorClient.Values: %v", err)
			return err
		}
		for _, v := range values {
			rates = append(rates, repo.RegistryRow{
				CurrencyPair: c,
				Time:         v.Time,
				Rate:         v.Rate,
			})
		}
	}

	if err = s.repo.Insert(ctx, rates); err != nil {
		s.logger.Error("Repo.Insert: %v", err)
		return err
	}

	return nil
}
