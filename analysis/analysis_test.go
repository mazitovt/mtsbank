package analysis

import (
	"context"
	"fmt"
	"math/rand"
	gs "mtsbank/analysis/internal/client/generator_service"
	hs "mtsbank/analysis/internal/client/history_service"
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
	v := []gs.ExchangeRateDTO{
		{
			Time: time.Now().Add(time.Duration(-30) * time.Second),
			Rate: int64(20 + rand.Int63n(10)),
		},
		{
			Time: time.Now().Add(time.Duration(-25) * time.Second),
			Rate: int64(30 + +rand.Int63n(10)),
		},
		{
			Time: time.Now(),
			Rate: int64(40 + +rand.Int63n(10)),
		},
	}
	for _, r := range v {
		fmt.Println(currencyPair, r)
	}
	return v, nil
}
func TestName(t *testing.T) {
	curPair := []string{"EURUSD"}
	timeFrames := []time.Duration{20 * time.Second}
	l := logger.New(logger.Debug)
	r := repo.NewInmemoryRepo(l)
	history, err := hs.NewClientWithResponses("http://localhost:8081")
	if err != nil {
		t.Fatal(err)
	}
	generator := gs.NewService("localhost:8080", l)

	resetPeriod := 1000 * time.Second
	pollPeriod := 5 * time.Second

	service := NewService(pollPeriod, curPair, timeFrames, generator, l, r)

	ctx := context.Background()
	service.Start(ctx, history, resetPeriod)

	select {}
}
