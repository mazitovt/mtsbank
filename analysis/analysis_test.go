package analysis

import (
	"context"
	"fmt"
	gs "mtsbank/analysis/internal/client/generator_service"
	"mtsbank/analysis/internal/repo"
	"mtsbank/analysis/logger"
	"testing"
	"time"
)

type fakeRepo struct {
}

func (f fakeRepo) Put(ctx context.Context, currencyPair string, timeFrame string, ohlc repo.OHLC) error {
	fmt.Println("fakeRepo.Put: ", currencyPair, timeFrame, ohlc)
	fmt.Printf("%+v", ohlc)
	return nil
}

func (f fakeRepo) Get(ctx context.Context, currencyPair string, timeFrame string) (*repo.OHLC, error) {
	//TODO implement me
	panic("implement me")
}

func (f fakeRepo) GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64) ([]repo.OHLC, error) {
	//TODO implement me
	panic("implement me")
}

func (f fakeRepo) Reset(ctx context.Context) {
	//TODO implement me
	panic("implement me")
}

type fakeGenerator struct {
}

func (f fakeGenerator) Values(ctx context.Context, currencyPair string) ([]gs.ExchangeRateDTO, error) {
	fmt.Println("fakeGenerator.Values", currencyPair)
	t := time.Now().Second() % 5
	fmt.Println(t)
	return []gs.ExchangeRateDTO{
		{
			Time: time.Now(),
			Rate: int64(t),
		},
		//{
		//	Time: time.Now().Add(time.Duration(-30) * time.Second),
		//	Rate: int64(20 + len(currencyPair)),
		//},
		//{
		//	Time: time.Now().Add(time.Duration(-25) * time.Second),
		//	Rate: int64(30 + len(currencyPair)),
		//},
		//{
		//	Time: time.Now(),
		//	Rate: int64(40 + len(currencyPair)),
		//},
	}, nil
}

func TestName(t *testing.T) {
	curPair := []string{"AB"}
	timeFrames := []time.Duration{3 * time.Second}
	l := logger.New(logger.Debug)
	repo := fakeRepo{}
	generator := fakeGenerator{}
	service := NewService(curPair, timeFrames, generator, l, repo)

	ctx := context.Background()
	resetPeriod := time.Hour
	pollPeriod := 1 * time.Second

	go service.StartTimeFrameGoroutines(ctx, resetPeriod)

	service.CollectNewRates(ctx, pollPeriod)
}
