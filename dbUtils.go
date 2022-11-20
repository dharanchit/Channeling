package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func connectAdventDb() *sql.DB {
	var (
		host     = os.Getenv("DATABASE_HOST")
		port     = os.Getenv("DATABASE_PORT")
		user     = os.Getenv("DATABASE_USER")
		password = os.Getenv("DATABASE_PASSWORD")
		dbname   = os.Getenv("DATABASE_NAME")
	)

	portInt, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, portInt, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	// defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}

func connectFundsDb() *sql.DB {
	var (
		host     = os.Getenv("FUNDS_DB_HOST")
		port     = os.Getenv("FUNDS_DB_PORT")
		user     = os.Getenv("FUNDS_DB_USER")
		password = os.Getenv("FUNDS_DB_PASSWORD")
		dbname   = os.Getenv("FUNDS_DB_NAME")
	)

	portInt, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, portInt, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}

type tickerPrices struct {
	price float64
	found bool
}

func getPriceFromAdvent(ticker string) map[string]tickerPrices {
	db := connectAdventDb()
	var (
		price float64
		found bool
	)

	result := make(map[string]tickerPrices)

	queryString := fmt.Sprintf("SELECT price FROM ticker_prices WHERE ticker = '%s';", ticker)

	rows, err := db.Query(queryString)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&price)
		found = true

		if err != nil {
			panic(err)
		}
	}
	result[ticker] = tickerPrices{price: price, found: found}
	return result

}

func getPriceFromFunds(ticker string) map[string]tickerPrices {
	db := connectFundsDb()
	var (
		closing_price float64
		found         bool
	)
	result := make(map[string]tickerPrices)

	queryString := fmt.Sprintf("SELECT closing_price FROM combined_ticker_prices WHERE ticker='%s';", ticker)
	rows, err := db.Query(queryString)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&closing_price)
		found = true
		if err != nil {
			panic(err)
		}
	}
	result[ticker] = tickerPrices{price: closing_price, found: found}
	return result
}

func getTickerPrices(ticker string, q chan tickerPriceResult) {
	adventResult := getPriceFromAdvent(ticker)
	if adventResult[ticker].found {
		q <- tickerPriceResult{ticker, adventResult[ticker].price}
	} else {
		fundsResult := getPriceFromFunds(ticker)
		if fundsResult[ticker].found {
			q <- tickerPriceResult{ticker, fundsResult[ticker].price}
		} else {
			q <- tickerPriceResult{ticker, nil}
		}
	}

}
