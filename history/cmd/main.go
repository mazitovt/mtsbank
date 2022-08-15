package main

import (
	"context"
	"fmt"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/go-chi/chi/v5"
	"github.com/mazitovt/logger"
	"log"
	"mtsbank/history"
	v1 "mtsbank/history/internal/api/http/v1"
	"mtsbank/history/internal/client/generator_service"
	"mtsbank/history/internal/config"
	"mtsbank/history/internal/repo"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {

	swagger, err := v1.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Create an instance of our handler which satisfies the generated interface
	cfg, err := config.Init()
	checkErr(err)

	l := logger.New(logger.Debug)

	repoPG, err := repo.NewRepoPG(&cfg.Postgres, l)
	if err != nil {
		log.Fatal("NewRepo ", err)
	}

	if cfg.Migrate {
		l.Debug("start migration")
		if err = repoPG.Migrate(); err != nil {
			log.Fatal(err)
		}
		l.Debug("end migration")
	}

	generatorClient := generator_service.NewService(net.JoinHostPort(cfg.Generator.Host, cfg.Generator.Port), l)

	g := history.NewSimpleHistoryService(repoPG, generatorClient, l)

	// This is how you set up a basic chi router
	r := chi.NewRouter()

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(middleware.OapiRequestValidator(swagger))

	// We now register our petStore above as the handler for the interface
	v1.HandlerFromMux(g, r)

	s := &http.Server{
		Handler: r,
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
	}

	// periodically call Collect
	go startCollecting(context.Background(), g, cfg.Period, l)
	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// TODO: encapsulate to HistoryService
func startCollecting(ctx context.Context, h history.HistoryService, period time.Duration, l logger.Logger) {

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		err := h.CollectExchangeRates(ctx)
		if err != nil {
			l.Error("HistoryService.CollectExchangeRates: %v", err)
		}

		select {
		case <-ctx.Done():
			fmt.Println("stop generating")
			return
		case <-ticker.C:
			continue
		}
	}
}
