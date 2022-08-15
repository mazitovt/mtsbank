package main

import (
	"context"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/go-chi/chi/v5"
	"github.com/mazitovt/logger"
	"log"
	"mtsbank/history/internal"
	"mtsbank/history/internal/api/http/v1"
	gs "mtsbank/history/internal/client/generator_service"
	"mtsbank/history/internal/config"
	"mtsbank/history/internal/repo"
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

	repoPG, err := repo.NewRepoPG(&cfg.Postgres, l)
	checkErr(err)

	if cfg.Migrate {
		checkErr(repoPG.Migrate())
	}

	genClient, err := gs.NewClientWithResponses("http://" + net.JoinHostPort(cfg.Generator.Host, cfg.Generator.Port))
	checkErr(err)

	service := internal.NewSimpleHistoryService(repoPG, genClient, l)

	// configure router
	swagger, err := v1.GetSwagger()
	checkErr(err)

	swagger.Servers = nil

	r := chi.NewRouter()
	r.Use(middleware.OapiRequestValidator(swagger))
	v1.HandlerFromMux(service, r)

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
	go service.Start(ctx, cfg.Period)
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
