package analyzer

import (
	"mtsbank/analysis/internal/model"
	"reflect"
	"testing"
	"time"
)

func Test_rateFilter_Check(t *testing.T) {
	type fields struct {
		last time.Time
	}
	type args struct {
		rates []model.ExchangeRate
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &rateFilter{
				last: tt.fields.last,
			}
			if got := tr.Check(tt.args.rates); got != tt.want {
				t.Errorf("rateFilter.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rateFilter_Last(t *testing.T) {
	type fields struct {
		last time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &rateFilter{
				last: tt.fields.last,
			}
			if got := tr.Last(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rateFilter.Last() = %v, want %v", got, tt.want)
			}
		})
	}
}
