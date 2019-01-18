package main

import (
	"encoding/json"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getAccountQuery(verbose, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf("%s", getAddressSpec(""))
		getAddress := getAddressClosure(cmd, "")

		cmd.Action = func() {
			address := getAddress()
			config := getConfig()
			ad, resp, err := tool.GetAccount(tmnode(config.Node, emitJSON, compact), address)
			if err != nil {
				finish(*verbose, resp, err, "account")
			}
			jsb, err := json.MarshalIndent(ad, "", "  ")
			fmt.Println(string(jsb))
			finish(*verbose, resp, err, "account")
		}
	}
}
