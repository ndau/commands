package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getAccountDelegate(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"NAME %s",
			getAddressSpec("NODE"),
		)

		var name = cmd.StringArg("NAME", "", "Name of account whose EAI calculations should be delegated")
		getNode := getAddressClosure(cmd, "NODE")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}
			if acct.Transfer == nil {
				orQuit(fmt.Errorf("Transfer key for %s not set", *name))
			}

			node := getNode()

			if *verbose {
				fmt.Printf(
					"Delegating %s to node %s\n",
					acct.Address.String(), node.String(),
				)
			}

			tx := ndau.NewDelegate(
				acct.Address, node,
				sequence(conf, acct.Address),
				acct.TransferPrivateK(keys),
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), tx)
			finish(*verbose, resp, err, "delegate")
		}
	}
}
