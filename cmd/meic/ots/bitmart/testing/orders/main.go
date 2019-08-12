package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"

	cli "github.com/jawher/mow.cli"
	bitmart "github.com/oneiro-ndev/commands/cmd/meic/ots/bitmart"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
)

func check(err error, context string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, context+":")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var currentIssued int64 = 0

func getCurrentIssued() int64 {
	if currentIssued == 0 {
		var pi routes.PriceInfo
		resp, err := http.Get("https://mainnet-0.ndau.tech:3030/price/current")
		check(err, "https get price/current")
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		check(err, "reading response body")
		var out bytes.Buffer
		err = json.Indent(&out, data, "", "  ")
		check(err, "formatting json")
		fmt.Printf("data = %s\n", out.Bytes())
		err = json.Unmarshal(data, &pi)
		check(err, "parsing price info response")
		fmt.Printf("current issued = %d", pi.TotalIssued)
		currentIssued = int64(pi.TotalIssued)
	}
	return currentIssued
}

func getBlockNum() int {
	return int(getCurrentIssued() / 100000000000)
}

func currentTargetPrice() float64 {
	var targPrice float64
	exp := (14 * float64(getBlockNum()) / float64(9999))
	targPrice = math.Pow(2, exp)
	return math.Round(targPrice*10000) / 10000
}

func getExchangeIssued(auth Auth, symbol *string) int64 {
	// get partial success orders, this will be the current stack level
	orders, err := bitmart.GetOrderHistory(&auth, *symbol, PartialSuccess)
	check(err, "getting orders")

	if len(orders) != 0 {
		exchangeIssued = calculateIssued(orders[0].Price, orders[0].ExecutedAmount)
	}

}
func main() {
	app := cli.App("orders", "get user orders from bitmart")

	var (
		apikeyPath = app.StringArg("API_KEY", "", "Path to an apikey.json file")
		symbol     = app.StringArg("SYMBOL", bitmart.NdauSymbol, "Trade symbol to examine")
		status     = app.StringArg("STATUS", bitmart.Invalid.String(), "order status filter")
		verbose    = app.BoolOpt("v verbose", false, "verbose mode")
	)

	app.Spec = "API_KEY [SYMBOL] STATUS [--verbose]"

	app.Action = func() {
		key, err := bitmart.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := bitmart.NewAuth(key)

		statusFilter := bitmart.OrderStatusFrom(*status)
		if *verbose {
			fmt.Println("using order status filter:", statusFilter)
		}
		orders, err := bitmart.GetOrderHistory(&auth, *symbol, statusFilter)
		check(err, "getting orders")

		exchangeIssued := getExchangeIssued(auth, symbol)

		data, err := json.MarshalIndent(orders, "", "  ")
		check(err, "formatting output")

		fmt.Println(string(data))
	}
	app.Run(os.Args)
}
