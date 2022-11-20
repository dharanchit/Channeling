package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/joho/godotenv"
)

type tickerPriceResult struct {
	ticker string
	price  interface{}
}

func tickerPrice(w http.ResponseWriter, r *http.Request) {
	result := make(map[string]map[string]interface{})

	if r.Method == "POST" {
		var payload map[string][]string
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			panic(err)
		}

		tickersList := payload["tickers"]

		for _, val := range tickersList {
			result[val] = make(map[string]interface{})
		}

		if len(tickersList) == 0 {
			w.Write([]byte("No tickers passed"))
		}

		var wg sync.WaitGroup

		q := make(chan tickerPriceResult)

		go func() {
			for res := range q {
				result[res.ticker]["closing_price"] = res.price
			}
		}()

		for _, val := range tickersList {
			wg.Add(1)

			go func(ticker string) {
				defer wg.Done()
				getTickerPrices(ticker, q)
			}(val)
		}
		wg.Wait()
		close(q)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonData, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		w.Write(jsonData)
	}

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	http.HandleFunc("/ticker-prices", tickerPrice)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
