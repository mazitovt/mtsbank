package generator_service

import (
	"context"
	"encoding/json"
	"github.com/mazitovt/logger"
	"net/http"
	"net/url"
)

type GeneratorService interface {
	Values(ctx context.Context, currencyPair string) ([]ExchangeRateDTO, error)
}

type client struct {
	url    string
	logger logger.Logger
}

func NewService(hostPort string, logger logger.Logger) GeneratorService {
	url := url.URL{
		Scheme: "http",
		Host:   hostPort,
		Path:   "rates",
	}
	return &client{
		url:    url.String(),
		logger: logger,
	}
}

func (c *client) Values(ctx context.Context, currencyPair string) ([]ExchangeRateDTO, error) {
	c.logger.Debug("client.Values(): %v", currencyPair)
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		c.logger.Error("http.NewRequest: ", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("currency_pair", currencyPair)
	req.URL.RawQuery = q.Encode()

	select {
	case <-ctx.Done():
		c.logger.Error("Context.Done: ", ctx.Err())
		return nil, ctx.Err()
	default:
	}

	c.logger.Debug(req.URL.String())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Error("Client.Do: ", err)
		return nil, err
	}

	defer resp.Body.Close()

	dto := []ExchangeRateDTO{}

	if err = json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		c.logger.Error("Decoder.Decode: ", err)
		return nil, err
	}

	c.logger.Debug("%v", dto)

	return dto, nil
}
