package main

import (
	"encoding/json"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/commands/cmd/bitmart"
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
		symbol     = app.StringArg("SYMBOL", "", "Trade symbol to examine")
	)

	app.Action = func() {
		key, err := bitmart.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := bitmart.NewAuth(key)
		trades, err := bitmart.GetTradeHistory(&auth, *symbol)
		check(err, "getting trades")

		data, err := json.MarshalIndent(trades, "", "  ")
		check(err, "formatting output")

		fmt.Println(string(data))
	}
	app.Run(os.Args)
}
