package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getSetRewardsDestination(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf("NAME %s", getAddressSpec("DESTINATION"))
		getAddress := getAddressClosure(cmd, "DESTINATION")
		var name = cmd.StringArg("NAME", "", "Name of account to lock")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}
			if acct.Transfer == nil {
				orQuit(fmt.Errorf("Transfer key for %s not set", *name))
			}

			dest := getAddress()

			if *verbose {
				fmt.Printf(
					"Setting rewards target for acct %s to %s\n",
					acct.Address,
					dest,
				)
			}

			tx := ndau.NewSetRewardsDestination(
				acct.Address,
				dest,
				sequence(conf, acct.Address),
				acct.TransferPrivateK(*keys)...,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), tx)
			finish(*verbose, resp, err, "set-rewards-target")
		}
	}
}
