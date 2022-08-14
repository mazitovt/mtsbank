package model

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var defaultTime = time.Time{}.Add(1 * time.Second)

func TestOHLC_UpdateOrReady(t *testing.T) {

	type fields struct {
		CurrencyPair string
		TimeFrame    time.Duration
		OpenTime     time.Time
		CloseTime    time.Time
		Open         int64
		High         int64
		Low          int64
		Close        int64
	}
	tests := []struct {
		name   string
		fields fields
		args   []ExchangeRate
		want   []bool
	}{
		{
			name: "last value out of time frame 1",
			fields: fields{
				TimeFrame: 5 * time.Second,
			},
			args: []ExchangeRate{
				{defaultTime, 5},
				{defaultTime.Add(2 * time.Second), 5},
				{defaultTime.Add(4 * time.Second), 10},
				{defaultTime.Add(6 * time.Second), 15},
			},
			want: []bool{false, false, false, true},
		},
		{
			name: "last value out of time frame 2",
			fields: fields{
				TimeFrame: 5 * time.Second,
			},
			args: []ExchangeRate{
				{defaultTime, 5},
				{defaultTime.Add(2 * time.Second), 5},
				{defaultTime.Add(4 * time.Second), 10},
				{defaultTime.Add(5001 * time.Millisecond), 15},
			},
			want: []bool{false, false, false, true},
		},
		{
			name: "all values are within time frame 1",
			fields: fields{
				TimeFrame: 5 * time.Second,
			},
			args: []ExchangeRate{
				{defaultTime, 5},
				{defaultTime.Add(2 * time.Second), 3},
				{defaultTime.Add(4 * time.Second), 10},
				{defaultTime.Add(5 * time.Second), 7},
			},
			want: []bool{false, false, false, false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OHLC{
				CurrencyPair: tt.fields.CurrencyPair,
				TimeFrame:    tt.fields.TimeFrame,
				OpenTime:     tt.fields.OpenTime,
				CloseTime:    tt.fields.CloseTime,
				Open:         tt.fields.Open,
				High:         tt.fields.High,
				Low:          tt.fields.Low,
				Close:        tt.fields.Close,
			}
			for i := range tt.args {
				if got := o.UpdateOrReady(tt.args[i]); got != tt.want[i] {
					t.Errorf("OHLC.UpdateOrReady() = %v, want %v", got, tt.want)
				}
				t.Log(o)
			}
		})
	}
}

func TestOHLC_Empty(t *testing.T) {
	type fields struct {
		CurrencyPair string
		TimeFrame    time.Duration
		OpenTime     time.Time
		CloseTime    time.Time
		Open         int64
		High         int64
		Low          int64
		Close        int64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "default value",
			want: true,
		},
		{
			name: "default value",
			fields: fields{
				OpenTime: defaultTime,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OHLC{
				CurrencyPair: tt.fields.CurrencyPair,
				TimeFrame:    tt.fields.TimeFrame,
				OpenTime:     tt.fields.OpenTime,
				CloseTime:    tt.fields.CloseTime,
				Open:         tt.fields.Open,
				High:         tt.fields.High,
				Low:          tt.fields.Low,
				Close:        tt.fields.Close,
			}
			if got := o.Empty(); got != tt.want {
				t.Errorf("OHLC.Empty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOHLC_Reset(t *testing.T) {
	type fields struct {
		CurrencyPair string
		TimeFrame    time.Duration
		OpenTime     time.Time
		CloseTime    time.Time
		Open         int64
		High         int64
		Low          int64
		Close        int64
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "reset 1",
			fields: fields{
				CurrencyPair: "ABC",
				TimeFrame:    5 * time.Second,
				OpenTime:     time.Now(),
				CloseTime:    time.Now().Add(5 * time.Second),
				Open:         1,
				High:         2,
				Low:          3,
				Close:        4,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OHLC{
				CurrencyPair: tt.fields.CurrencyPair,
				TimeFrame:    tt.fields.TimeFrame,
				OpenTime:     tt.fields.OpenTime,
				CloseTime:    tt.fields.CloseTime,
				Open:         tt.fields.Open,
				High:         tt.fields.High,
				Low:          tt.fields.Low,
				Close:        tt.fields.Close,
			}
			o.Reset()
			require.Equal(t, o.CurrencyPair, tt.fields.CurrencyPair)
			require.Equal(t, o.TimeFrame, tt.fields.TimeFrame)
			require.Equal(t, o.OpenTime, emptyTime)
			require.Equal(t, o.CloseTime, emptyTime)
			require.Equal(t, o.Open, emptyRate)
			require.Equal(t, o.High, emptyRate)
			require.Equal(t, o.Low, emptyRate)
			require.Equal(t, o.Close, emptyRate)
		})
	}
}
