package v1

import (
	"context"
	"errors"
	api "mtsbank/history/internal/api/http/v1"
)

var _ GeneratorService = (*ClientWithResponses)(nil)

var (
	ErrNoDecodedValues = errors.New("no decoded values")
	ErrBufferGrow      = errors.New("buffer had been grown")
)

type GeneratorService interface {
	GetRates(ctx context.Context, currencyPair string, out []api.ExchangeRate) ([]api.ExchangeRate, error)
}

// GetRates grows out slice and copies new rates to out slice.
//
// Makes a blocking http call
func (c *ClientWithResponses) GetRates(ctx context.Context, currencyPair string, buffer []api.ExchangeRate) ([]api.ExchangeRate, error) {
	resp, err := c.GetRatesCurrencyPairWithResponse(ctx, currencyPair)
	if err != nil {
		return buffer, err
	}

	if resp.JSON200 == nil {
		return buffer, ErrNoDecodedValues
	}

	err = nil

	if len(*resp.JSON200) > cap(buffer) {
		buffer = make([]api.ExchangeRate, 0, len(*resp.JSON200))
		err = ErrBufferGrow
	}

	buffer = buffer[:len(*resp.JSON200)]

	for i := range buffer {
		buffer[i] = api.ExchangeRate{
			Rate: (*resp.JSON200)[i].Rate,
			Time: (*resp.JSON200)[i].Time,
		}
	}

	return buffer, err
}
