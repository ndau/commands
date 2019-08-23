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
	bitmart "github.com/oneiro-ndev/commands/cmd/bitmart/api"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
)

func check(err error, context string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, context+":")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func targetPrice(blockNum int) float64 {
	var targPrice float64
	exp := (14 * float64(blockNum)) / float64(9999)
	targPrice = math.Pow(2, exp)
	return math.Round(targPrice*10000) / 10000
}

func getBlockNum() int {
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
	return int(pi.TotalIssued) / 100000000000
}

func main() {
	app := cli.App("trades", "get user trades from bitmart")

	var (
		apikeyPath = app.StringArg("API_KEY", "", "Path to an apikey.json file")
		symbol     = app.StringArg("SYMBOL", bitmart.NdauSymbol, "Trade symbol to examine")
		limit      = app.IntOpt("limit", 0, "return only values with trade_id > limit")
	)

	app.Spec = "API_KEY [SYMBOL] [--limit]"

	blockNum := getBlockNum()

	app.Action = func() {
		key, err := bitmart.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := bitmart.NewAuth(key)
		trades, err := bitmart.GetTradeHistoryAfter(&auth, *symbol, int64(*limit))
		check(err, "getting trades")

		data, err := json.MarshalIndent(trades, "", "  ")
		check(err, "formatting output")

		fmt.Println(string(data))
		for block := blockNum; block < blockNum+100; block++ {
			fmt.Printf("target price = %f\n", targetPrice(block))
		}
	}
	app.Run(os.Args)
}
