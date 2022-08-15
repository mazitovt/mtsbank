package main

import (
	"context"
	"encoding/json"
	"fmt"
	"generator/internal"
	"generator/internal/config"
	"github.com/mazitovt/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	cfg, err := config.Init()
	checkErr(err)

	f, err := config.GetGeneratorFunc(cfg)
	checkErr(err)

	g := internal.NewSimplePriceGenerator(cfg.CurrencyPairs, f, uint64(cfg.CacheSize), logger.New(logger.Debug))

	ctx, cancel := context.WithCancel(context.Background())

	// gracefully shutdown when signal comes
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("cancel signal")
		cancel()
	}()

	// periodically call Generate
	go startGenerate(ctx, g, cfg.Period)

	// wait for incoming requests
	http.HandleFunc("/rates", newHandler(g))
	s := http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	go func() {
		<-ctx.Done()
		fmt.Println("server shutdown")
		if err := s.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
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

func startGenerate(ctx context.Context, g internal.PriceGenerator, period time.Duration) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("stop generating")
			return
		default:
			g.Generate()
			time.Sleep(period)
		}
	}
}

func newHandler(g internal.PriceGenerator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !r.URL.Query().Has("currency_pair") {
			w.Write([]byte("parameter 'currency_pair' is missing"))
			return
		}

		pair := r.URL.Query().Get("currency_pair")

		get, err := g.Get(pair)
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
