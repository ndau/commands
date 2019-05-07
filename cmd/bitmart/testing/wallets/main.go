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
	app := cli.App("wallets", "get user wallets from bitmart")

	var apikeyPath = app.StringArg("API_KEY", "", "Path to an apikey.json file")

	app.Action = func() {
		key, err := bitmart.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := bitmart.NewAuth(key)
		wallets, err := bitmart.GetWallets(&auth)
		check(err, "getting wallets")

		data, err := json.MarshalIndent(wallets, "", "  ")
		check(err, "formatting output")

		fmt.Println(string(data))
	}
	app.Run(os.Args)
}
