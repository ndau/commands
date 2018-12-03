package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getStake(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"NAME %s",
			getAddressSpec("NODE"),
		)

		var name = cmd.StringArg("NAME", "", "Name of account to stake")
		getNode := getAddressClosure(cmd, "NODE")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}
			if len(acct.Transfer) == 0 {
				orQuit(fmt.Errorf("Transfer key for %s not set", *name))
			}

			node := getNode()

			if *verbose {
				fmt.Printf(
					"Staking acct %s to %s\n",
					acct.Address,
					node,
				)
			}

			tx := ndau.NewStake(
				acct.Address, node,
				sequence(conf, acct.Address),
				acct.TransferPrivateK(keys),
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), tx)
			finish(*verbose, resp, err, "notify")
		}
	}
}
