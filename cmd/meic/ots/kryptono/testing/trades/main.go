package main

import (
	"encoding/json"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	kryptono "github.com/oneiro-ndev/commands/cmd/meic/ots/kryptono"
)

func check(err error, context string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, context+":")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	app := cli.App("trades", "get user trades from kryptono")

	var (
		apikeyPath = app.StringArg("API_KEY", "", "Path to an apikey.json file")
		symbol     = app.StringArg("SYMBOL", "XND_USDT", "Trade symbol to examine")
		limit      = app.StringArg("LIMIT", "", "return only values with trade_id > limit")
	)

	app.Spec = "API_KEY [SYMBOL] [LIMIT]"

	app.Action = func() {
		key, err := kryptono.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := kryptono.NewAuth(key)
		trades, max, err := kryptono.GetTradeHistoryAfter(&auth, *symbol, *limit)
		check(err, "getting trades")

		data, err := json.MarshalIndent(trades, "", "  ")
		check(err, "formatting output")

		fmt.Println(string(data))
		fmt.Println("MaxID = ", max)
	}
	app.Run(os.Args)
}
