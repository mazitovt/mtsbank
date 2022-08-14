package analysis

import (
	"context"
	"fmt"
	"math/rand"
	gs "mtsbank/analysis/internal/client/generator_service"
	hs "mtsbank/analysis/internal/client/history_service"
	"mtsbank/analysis/internal/model"
	"mtsbank/analysis/internal/repo"
	"mtsbank/analysis/logger"
	"testing"
	"time"
)

type fakeRepo struct {
}

func (f fakeRepo) Put(ctx context.Context, currencyPair string, timeFrame string, ohlc model.OHLC) error {
	fmt.Println("fakeRepo.Put: ", currencyPair, timeFrame, ohlc)
	fmt.Printf("%+v", ohlc)
	return nil
}

func (f fakeRepo) Get(ctx context.Context, currencyPair string, timeFrame string) (*model.OHLC, error) {
	//TODO implement me
	panic("implement me")
}

func (f fakeRepo) GetMany(ctx context.Context, currencyPair string, timeFrame string, last int64) ([]model.OHLC, error) {
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

func TestName1(t *testing.T) {
	curPair := []string{"EURUSD"}
	timeFrames := []time.Duration{5 * time.Second, 10 * time.Second}
	l := logger.New(logger.Info)
	r := repo.NewInmemoryRepo(l)
	history, err := hs.NewClientWithResponses("http://localhost:8081")
	if err != nil {
		t.Fatal(err)
	}
	generator := gs.NewService("localhost:8080", l)

	resetPeriod := 30 * time.Second
	pollPeriod := 30 * time.Second
	batchPeriod := 30 * time.Second
	batchSize := 1

	service := NewService(pollPeriod, curPair, timeFrames, generator, l, r)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
		fmt.Println("CONTEXT IS DONE")
	}()

	service.Start(ctx, history, resetPeriod, batchPeriod, batchSize)

	for c, tm := range r.GetAll() {
		fmt.Println(c)
		if tm != nil {
			for tf, s := range tm {
				fmt.Println(tf)
				for _, v := range s {
					fmt.Println(v)
				}
			}
		}
	}
}
