package main

import (
	"encoding/json"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	bitmart "github.com/oneiro-ndev/commands/cmd/meic/ots/bitmart"
)

func check(err error, context string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, context+":")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	app := cli.App("trades", "get user trades from bitmart")

	var (
		apikeyPath = app.StringArg("API_KEY", "", "Path to an apikey.json file")
		symbol     = app.StringArg("SYMBOL", bitmart.NdauSymbol, "Trade symbol to examine")
		limit      = app.IntOpt("limit", 0, "return only values with trade_id > limit")
	)

	app.Spec = "API_KEY [SYMBOL] [--limit]"

	app.Action = func() {
		key, err := bitmart.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := bitmart.NewAuth(key)
		trades, _, err := bitmart.GetTradeHistoryAfter(&auth, *symbol, int64(*limit))
		check(err, "getting trades")

		data, err := json.MarshalIndent(trades, "", "  ")
		check(err, "formatting output")

		fmt.Println(string(data))
	}
	app.Run(os.Args)
}
