package history

import (
	"context"
	generator "mtsbank/history/internal/client/generator_service"
	"mtsbank/history/internal/repo"
	"mtsbank/history/logger"
	"time"
)

type ExchangeRate struct {
	Time time.Time
	Rate int64
}

type HistoryService interface {
	GetByTime(ctx context.Context, currencyPair string, start time.Time, end time.Time) ([]ExchangeRate, error)
	CollectExchangeRates(ctx context.Context) error
}

type SimpleHistoryService struct {
	repo            repo.Repo
	generatorClient generator.GeneratorService
	logger          logger.Logger
}

func NewSimpleHistoryService(repo repo.Repo, generatorClient generator.GeneratorService, logger logger.Logger) *SimpleHistoryService {
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
