package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getAccountList(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command("remote", "list accounts known to the ndau chain", getAccountListRemote(verbose))

		cmd.Action = func() {
			config := getConfig()
			config.EmitAccounts(os.Stdout)
		}
	}
}

func getAccountListRemote(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			accts, err := tool.GetAccountListBatch(tmnode(config.Node, nil, nil))
			orQuit(err)

			for _, acct := range accts {
				fmt.Println(acct)
			}
			finish(*verbose, nil, err, "account list remote")
		}
	}
}
