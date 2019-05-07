package main

import (
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
	app := cli.App("auth_rest", "perform REST API authentication for bitmart")

	var apikeyPath = app.StringArg("API_KEY", "", "Path to an apikey.json file")

	app.Action = func() {
		key, err := bitmart.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")

		token, err := key.Authenticate()
		check(err, "authenticating")

		fmt.Printf("%#v\n", token)
	}
	app.Run(os.Args)
}
