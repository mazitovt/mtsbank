package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mtsbank/history"
	"mtsbank/history/internal/client/generator_service"
	"mtsbank/history/internal/config"
	"mtsbank/history/internal/repo"
	"mtsbank/history/logger"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Init()
	checkErr(err)

	l := logger.New(logger.Debug)

	repo, err := repo.NewRepoPG(&cfg.Postgres, l)
	if err != nil {
		log.Fatal("NewRepo ", err)
	}

	if cfg.Migrate {
		l.Debug("start migration")
		if err = repo.Migrate(); err != nil {
			log.Fatal(err)
		}
		l.Debug("end migration")
	}

	generatorClient := generator_service.NewService(net.JoinHostPort(cfg.Generator.Host, cfg.Generator.Port), l)

	g := history.NewSimpleHistoryService(repo, generatorClient, l)

	ctx, cancel := context.WithCancel(context.Background())

	// gracefully shutdown when signal comes
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("cancel signal")
		cancel()
	}()

	// periodically call Collect
	go startCollecting(ctx, g, cfg.Period, l)

	// wait for incoming requests
	http.HandleFunc("/rates", newHandler(g))
	s := http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: http.DefaultServeMux,
	}

	go func() {
		<-ctx.Done()
		fmt.Println("server shutdown")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	err = s.ListenAndServe()
	time.Sleep(cfg.Period)
	log.Fatal(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func startCollecting(ctx context.Context, h history.HistoryService, period time.Duration, l logger.Logger) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("stop generating")
			return
		default:
			err := h.CollectExchangeRates(ctx)
			if err != nil {
				l.Error("HistoryService.CollectExchangeRates: %v", err)
				l.Info("stop collecting")
				return
			}
			time.Sleep(period)
		}
	}
}

func newHandler(h history.HistoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !r.URL.Query().Has("currency_pair") {
			w.Write([]byte("parameter 'currency_pair' is missing"))
			return
		}

		if !r.URL.Query().Has("from") {
			w.Write([]byte("parameter 'from' is missing"))
			return
		}

		if !r.URL.Query().Has("to") {
			w.Write([]byte("parameter 'to' is missing"))
			return
		}

		pair := r.URL.Query().Get("currency_pair")
		fromParam := r.URL.Query().Get("from")
		toParam := r.URL.Query().Get("to")

		from, err := time.Parse(time.RFC3339, fromParam)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		to, err := time.Parse(time.RFC3339, toParam)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		get, err := h.GetByTime(r.Context(), pair, from, to)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(get)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
	}
}

//if err := srv.ListenAndServe(); err != http.ErrServerClosed {
//// Error starting or closing listener:
//log.Fatalf("HTTP server ListenAndServe: %v", err)
//}
