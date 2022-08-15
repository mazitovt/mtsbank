package analyzer

import (
	"github.com/mazitovt/logger"
	"mtsbank/analysis/internal/model"
	"reflect"
	"testing"
	"time"
)

func TestNewCurrencyPairAnalyzer(t *testing.T) {
	type args struct {
		currencyPair string
		timeFrames   []time.Duration
		logger       logger.Logger
	}
	tests := []struct {
		name string
		args args
		want *CurrencyPairAnalyzer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCurrencyPairAnalyzer(tt.args.currencyPair, tt.args.timeFrames, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCurrencyPairAnalyzer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyPairAnalyzer_CurrencyPair(t *testing.T) {
	type fields struct {
		timeFrames   []time.Duration
		currencyPair string
		logger       logger.Logger
		rateFilter   rateFilter
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CurrencyPairAnalyzer{
				timeFrames:   tt.fields.timeFrames,
				currencyPair: tt.fields.currencyPair,
				logger:       tt.fields.logger,
				rateFilter:   tt.fields.rateFilter,
			}
			if got := c.CurrencyPair(); got != tt.want {
				t.Errorf("CurrencyPairAnalyzer.CurrencyPair() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyPairAnalyzer_Start(t *testing.T) {
	type fields struct {
		timeFrames   []time.Duration
		currencyPair string
		logger       logger.Logger
		rateFilter   rateFilter
	}
	type args struct {
		in <-chan []model.ExchangeRate
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   <-chan model.OHLC
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CurrencyPairAnalyzer{
				timeFrames:   tt.fields.timeFrames,
				currencyPair: tt.fields.currencyPair,
				logger:       tt.fields.logger,
				rateFilter:   tt.fields.rateFilter,
			}
			if got := c.Start(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CurrencyPairAnalyzer.Start() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyPairAnalyzer_receive(t *testing.T) {
	type fields struct {
		timeFrames   []time.Duration
		currencyPair string
		logger       logger.Logger
		rateFilter   rateFilter
	}
	type args struct {
		in   <-chan []model.ExchangeRate
		outs []chan<- []model.ExchangeRate
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
			c := &CurrencyPairAnalyzer{
				timeFrames:   tt.fields.timeFrames,
				currencyPair: tt.fields.currencyPair,
				logger:       tt.fields.logger,
				rateFilter:   tt.fields.rateFilter,
			}
			c.receive(tt.args.in, tt.args.outs)
		})
	}
}

func TestCurrencyPairAnalyzer_analyze(t *testing.T) {
	type fields struct {
		timeFrames   []time.Duration
		currencyPair string
		logger       logger.Logger
		rateFilter   rateFilter
	}
	type args struct {
		in        <-chan []model.ExchangeRate
		out       chan<- model.OHLC
		timeFrame time.Duration
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
			c := &CurrencyPairAnalyzer{
				timeFrames:   tt.fields.timeFrames,
				currencyPair: tt.fields.currencyPair,
				logger:       tt.fields.logger,
				rateFilter:   tt.fields.rateFilter,
			}
			c.analyze(tt.args.in, tt.args.out, tt.args.timeFrame)
		})
	}
}
