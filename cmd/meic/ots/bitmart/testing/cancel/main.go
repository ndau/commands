package main

import (
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
	app := cli.App("cancel", "cancel order from bitmart")

	var (
		apikeyPath = app.StringArg("API_KEY", "", "Path to an apikey.json file")
		order      = app.IntOpt("order", 0, "cancel order ID ")
	)

	app.Spec = "API_KEY [--order]"

	app.Action = func() {
		key, err := bitmart.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := bitmart.NewAuth(key)
		fmt.Println("order = ", *order)
		err = bitmart.CancelOrder(&auth, uint64(*order))
		check(err, "cancel order")
	}
	app.Run(os.Args)
}
