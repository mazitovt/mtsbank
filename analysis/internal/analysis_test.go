package internal

import (
	"context"
	"fmt"
	"github.com/mazitovt/logger"
	"math/rand"
	"mtsbank/analysis/internal/analyzer"
	gs "mtsbank/analysis/internal/client/generator_service"
	hs "mtsbank/analysis/internal/client/history_service"
	"mtsbank/analysis/internal/model"
	"mtsbank/analysis/internal/repo"
	"reflect"
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

//func TestName1(t *testing.T) {
//	curPair := []string{"EURUSD"}
//	timeFrames := []time.Duration{5 * time.Second, 10 * time.Second}
//	l := logger.New(logger.Info)
//	r := repo.NewInmemoryRepo(l)
//	history, err := hs.NewClientWithResponses("http://localhost:8081")
//	if err != nil {
//		t.Fatal(err)
//	}
//	generator := gs.NewService("localhost:8080", l)
//
//	resetPeriod := 30 * time.Second
//	pollPeriod := 30 * time.Second
//	batchPeriod := 30 * time.Second
//	batchSize := 1
//
//	service := NewService(pollPeriod, curPair, timeFrames, generator, l, r)
//
//	ctx, cancel := context.WithCancel(context.Background())
//
//	go func() {
//		time.Sleep(1 * time.Second)
//		cancel()
//		fmt.Println("CONTEXT IS DONE")
//	}()
//
//	service.Start(ctx, history, resetPeriod, batchPeriod, batchSize)
//
//	for c, tm := range r.GetAll() {
//		fmt.Println(c)
//		if tm != nil {
//			for tf, s := range tm {
//				fmt.Println(tf)
//				for _, v := range s {
//					fmt.Println(v)
//				}
//			}
//		}
//	}
//}

func TestStart2(t *testing.T) {
	l := logger.New(logger.Info)

	curPair := []string{"EURUSD"}
	timeFrames := []time.Duration{5 * time.Second}

	analyzers := make([]analyzer.Analyzer, len(curPair))
	for i := range analyzers {
		analyzers[i] = analyzer.NewCurrencyPairAnalyzer(curPair[i], timeFrames, l)
	}

	r := repo.NewInmemoryRepo(l)
	history, err := hs.NewClientWithResponses("http://localhost:8081")
	if err != nil {
		t.Fatal(err)
	}
	generator := gs.NewService("localhost:8080", l)

	resetPeriod := 30 * time.Second
	pollPeriod := 1 * time.Second
	batchPeriod := 30 * time.Second
	batchSize := 5

	service := NewService(analyzers, batchPeriod, batchSize, resetPeriod, pollPeriod, history, generator, r, l)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(7 * time.Second)
		cancel()
		fmt.Println("CONTEXT IS DONE")
	}()

	service.Start(ctx)

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

func TestNewService(t *testing.T) {
	type args struct {
		currencyPairAnalyzers []analyzer.Analyzer
		batchPeriod           time.Duration
		batchSize             int
		resetPeriod           time.Duration
		pollPeriod            time.Duration
		history               hs.HistoryService
		generator             gs.GeneratorService
		repo                  repo.Repo
		logger                logger.Logger
	}
	tests := []struct {
		name string
		args args
		want *service
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewService(tt.args.currencyPairAnalyzers, tt.args.batchPeriod, tt.args.batchSize, tt.args.resetPeriod, tt.args.pollPeriod, tt.args.history, tt.args.generator, tt.args.repo, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_Start(t *testing.T) {
	type fields struct {
		currencyPairAnalyzers []analyzer.Analyzer
		batchPeriod           time.Duration
		batchSize             int
		resetPeriod           time.Duration
		pollPeriod            time.Duration
		history               hs.HistoryService
		generator             gs.GeneratorService
		repo                  repo.Repo
		logger                logger.Logger
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				currencyPairAnalyzers: tt.fields.currencyPairAnalyzers,
				batchPeriod:           tt.fields.batchPeriod,
				batchSize:             tt.fields.batchSize,
				resetPeriod:           tt.fields.resetPeriod,
				pollPeriod:            tt.fields.pollPeriod,
				history:               tt.fields.history,
				generator:             tt.fields.generator,
				repo:                  tt.fields.repo,
				logger:                tt.fields.logger,
			}
			s.Start(tt.args.ctx)
		})
	}
}

func Test_service_newCycle2(t *testing.T) {
	type fields struct {
		currencyPairAnalyzers []analyzer.Analyzer
		batchPeriod           time.Duration
		batchSize             int
		resetPeriod           time.Duration
		pollPeriod            time.Duration
		history               hs.HistoryService
		generator             gs.GeneratorService
		repo                  repo.Repo
		logger                logger.Logger
	}
	type args struct {
		ctx              context.Context
		historyCollected <-chan struct{}
		a                analyzer.Analyzer
		h                hs.HistoryService
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				currencyPairAnalyzers: tt.fields.currencyPairAnalyzers,
				batchPeriod:           tt.fields.batchPeriod,
				batchSize:             tt.fields.batchSize,
				resetPeriod:           tt.fields.resetPeriod,
				pollPeriod:            tt.fields.pollPeriod,
				history:               tt.fields.history,
				generator:             tt.fields.generator,
				repo:                  tt.fields.repo,
				logger:                tt.fields.logger,
			}
			s.newCycle2(tt.args.ctx, tt.args.historyCollected, tt.args.a, tt.args.h)
		})
	}
}

func Test_service_collectHistory(t *testing.T) {
	type fields struct {
		currencyPairAnalyzers []analyzer.Analyzer
		batchPeriod           time.Duration
		batchSize             int
		resetPeriod           time.Duration
		pollPeriod            time.Duration
		history               hs.HistoryService
		generator             gs.GeneratorService
		repo                  repo.Repo
		logger                logger.Logger
	}
	type args struct {
		ctx          context.Context
		in           chan<- []model.ExchangeRate
		currencyPair string
		h            hs.HistoryService
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				currencyPairAnalyzers: tt.fields.currencyPairAnalyzers,
				batchPeriod:           tt.fields.batchPeriod,
				batchSize:             tt.fields.batchSize,
				resetPeriod:           tt.fields.resetPeriod,
				pollPeriod:            tt.fields.pollPeriod,
				history:               tt.fields.history,
				generator:             tt.fields.generator,
				repo:                  tt.fields.repo,
				logger:                tt.fields.logger,
			}
			s.collectHistory(tt.args.ctx, tt.args.in, tt.args.currencyPair, tt.args.h)
		})
	}
}

func Test_service_collectNewRates(t *testing.T) {
	type fields struct {
		currencyPairAnalyzers []analyzer.Analyzer
		batchPeriod           time.Duration
		batchSize             int
		resetPeriod           time.Duration
		pollPeriod            time.Duration
		history               hs.HistoryService
		generator             gs.GeneratorService
		repo                  repo.Repo
		logger                logger.Logger
	}
	type args struct {
		ctx          context.Context
		in           chan<- []model.ExchangeRate
		currencyPair string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				currencyPairAnalyzers: tt.fields.currencyPairAnalyzers,
				batchPeriod:           tt.fields.batchPeriod,
				batchSize:             tt.fields.batchSize,
				resetPeriod:           tt.fields.resetPeriod,
				pollPeriod:            tt.fields.pollPeriod,
				history:               tt.fields.history,
				generator:             tt.fields.generator,
				repo:                  tt.fields.repo,
				logger:                tt.fields.logger,
			}
			s.collectNewRates(tt.args.ctx, tt.args.in, tt.args.currencyPair)
		})
	}
}

func Test_service_storeOHLC(t *testing.T) {
	type fields struct {
		currencyPairAnalyzers []analyzer.Analyzer
		batchPeriod           time.Duration
		batchSize             int
		resetPeriod           time.Duration
		pollPeriod            time.Duration
		history               hs.HistoryService
		generator             gs.GeneratorService
		repo                  repo.Repo
		logger                logger.Logger
	}
	type args struct {
		in          <-chan model.OHLC
		batchPeriod time.Duration
		batchSize   int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				currencyPairAnalyzers: tt.fields.currencyPairAnalyzers,
				batchPeriod:           tt.fields.batchPeriod,
				batchSize:             tt.fields.batchSize,
				resetPeriod:           tt.fields.resetPeriod,
				pollPeriod:            tt.fields.pollPeriod,
				history:               tt.fields.history,
				generator:             tt.fields.generator,
				repo:                  tt.fields.repo,
				logger:                tt.fields.logger,
			}
			s.storeOHLC(tt.args.in, tt.args.batchPeriod, tt.args.batchSize)
		})
	}
}