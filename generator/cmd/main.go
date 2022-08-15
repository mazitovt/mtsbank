package main

import (
	"context"
	"generator/internal"
	"generator/internal/api/http/v1"
	"generator/internal/config"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/go-chi/chi/v5"
	"github.com/mazitovt/logger"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const defaultLogLevel = logger.Info

func main() {

	// configure service
	cfg, err := config.Init()
	checkErr(err)

	var l logger.Logger
	if level, err := logger.LevelFromString(cfg.LogLevel); err != nil {
		l = logger.New(defaultLogLevel)
		l.Warn("unknown log level. set log level to default")
	} else {
		l = logger.New(level)
	}

	f, err := config.GetGeneratorFunc(cfg)
	checkErr(err)

	g := internal.NewSimplePriceGenerator(cfg.CurrencyPairs, f, uint64(cfg.CacheSize), l)

	// configure router
	swagger, err := v1.GetSwagger()
	checkErr(err)

	swagger.Servers = nil

	r := chi.NewRouter()
	r.Use(middleware.OapiRequestValidator(swagger))
	v1.HandlerFromMux(g, r)

	s := &http.Server{
		Handler: r,
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
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
	go g.Start(ctx, cfg.Period)
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
