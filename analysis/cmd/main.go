package main

import (
	"context"
	"fmt"
	"log"
	hs "mtsbank/analysis/internal/client/history_service"
	"time"
)

func main() {

	client, err := hs.NewClientWithResponses("http://localhost:8081")
	if err != nil {
		log.Fatal(err)
	}

	t1 := time.Now().Truncate(time.Hour)
	t2 := time.Now()
	resp, err := client.GetRatesCurrencyPairWithResponse(context.Background(), "EURUSD", &hs.GetRatesCurrencyPairParams{
		From: &t1,
		To:   &t2,
	})
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(resp.HTTPResponse)
	//fmt.Println(resp.StatusCode())
	//fmt.Println(resp.HTTPResponse)
	fmt.Println(resp.JSON200)
	//fmt.Println(string(resp.Body))
	//fmt.Println(resp.JSON200)
	//fmt.Println(resp.JSONDefault)
}
