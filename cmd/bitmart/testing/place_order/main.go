package main

import (
	"encoding/json"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	bitmart "github.com/oneiro-ndev/commands/cmd/bitmart/api"
)

func check(err error, context string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, context+":")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	app := cli.App("place_order", "place an ndau order on bitmart")

	var (
		apikeyPath = app.StringArg("API_KEY", "", "path to an apikey.json file")
		side       = app.StringArg("SIDE", "", "\"buy\" or \"sell\"")
		symbol     = app.StringArg("SYMBOL", bitmart.NdauSymbol, "symbol to trade")
		amount     = app.Float64Arg("AMOUNT", 0, "qty desired")
		price      = app.Float64Arg("PRICE", 0, "price in denominated unit")
	)

	app.Spec = "API_KEY SIDE [SYMBOL] AMOUNT PRICE"

	app.Action = func() {
		key, err := bitmart.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := bitmart.NewAuth(key)

		resp, err := bitmart.PlaceOrder(&auth, *symbol, *side, *price, *amount)
		check(err, "placing order")

		output, err := json.MarshalIndent(resp, "", "  ")
		check(err, "formatting output")

		fmt.Println(output)
	}

	app.Run(os.Args)
}
