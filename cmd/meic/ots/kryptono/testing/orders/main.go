package main

import (
	"encoding/json"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	// "bytes"
	// "encoding/json"
	// "fmt"
	// "io/ioutil"
	// "math"
	// "net/http"
	// "os"
	// cli "github.com/jawher/mow.cli"
	kryptono "github.com/oneiro-ndev/commands/cmd/meic/ots/kryptono"
	// "github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
)

func check(err error, context string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, context+":")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	app := cli.App("orders", "get user orders from bitmart")

	var (
		apikeyPath = app.StringArg("API_KEY", "", "Path to an apikey.json file")
		symbol     = app.StringArg("SYMBOL", "", "Trade symbol to examine")
		status     = app.StringArg("STATUS", "", "order status filter")
		verbose    = app.BoolOpt("v verbose", false, "verbose mode")
	)

	app.Spec = "API_KEY [SYMBOL] STATUS [--verbose]"

	app.Action = func() {
		key, err := kryptono.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := kryptono.NewAuth(key)

		statusFilter := kryptono.OrderStatusFrom(*status)
		if *verbose {
			fmt.Println("using order status filter:", statusFilter)
		}
		orders, err := kryptono.GetOrderHistory(&auth, *symbol, statusFilter)
		check(err, "getting orders")

		data, err := json.MarshalIndent(orders, "", "  ")
		check(err, "formatting output")

		fmt.Println(string(data))
	}
	app.Run(os.Args)
}
