package main

import (
	"context"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/go-chi/chi/v5"
	"github.com/mazitovt/logger"
	"log"
	"mtsbank/analysis/internal"
	"mtsbank/analysis/internal/analyzer"
	"mtsbank/analysis/internal/api/http/v1"
	gs "mtsbank/analysis/internal/client/generator_service"
	hs "mtsbank/analysis/internal/client/history_service"
	"mtsbank/analysis/internal/config"
	"mtsbank/analysis/internal/repo"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const defaultLogLevel = logger.Info

func main() {

	// Create an instance of our handler which satisfies the generated interface
	cfg, err := config.Init()
	checkErr(err)

	var l logger.Logger
	if level, err := logger.LevelFromString(cfg.LogLevel); err != nil {
		l = logger.New(defaultLogLevel)
		l.Warn("unknown log level. set log level to default")
	} else {
		l = logger.New(level)
	}

	analyzers := make([]analyzer.Analyzer, len(cfg.CurrencyPairs))
	for i := range analyzers {
		analyzers[i] = analyzer.NewCurrencyPairAnalyzer(cfg.CurrencyPairs[i], cfg.TimeFrames, l)
	}

	memRepo := repo.NewInmemoryRepo(l)

	l.Debug("%+v", cfg.Generator)
	history, err := hs.NewClientWithResponses("http://" + net.JoinHostPort(cfg.History.Host, cfg.History.Port))
	checkErr(err)

	l.Debug("%+v", cfg.History)
	generator, err := gs.NewClientWithResponses("http://" + net.JoinHostPort(cfg.Generator.Host, cfg.Generator.Port))
	checkErr(err)

	service := internal.NewService(
		analyzers,
		cfg.Batch.Period,
		int(cfg.Batch.Size),
		cfg.RestartAfter,
		cfg.PollPeriod,
		history,
		generator,
		memRepo,
		l)

	// configure router
	swagger, err := v1.GetSwagger()
	checkErr(err)

	swagger.Servers = nil

	r := chi.NewRouter()
	r.Use(middleware.OapiRequestValidator(swagger))
	v1.HandlerFromMux(service, r)

	s := &http.Server{
		Handler: r,
		Addr:    net.JoinHostPort(cfg.Http.Host, cfg.Http.Port),
	}

	// shutdown gracefully
	ctx, cancel := context.WithCancel(context.Background())

	idleConnsClosed := make(chan struct{})

	exit := make(chan os.Signal, 1)

	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-exit
		cancel()
		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	// Start service
	go service.Start(ctx)
	l.Info("Service started")

	// Start server
	if err = s.ListenAndServe(); err != http.ErrServerClosed {
		checkErr(err)
	}

	<-idleConnsClosed

	l.Info("Service stopped")
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
