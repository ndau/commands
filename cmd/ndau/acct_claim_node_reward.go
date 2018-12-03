package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getClaimNodeReward(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME"

		var name = cmd.StringArg("NAME", "", "Name of account to lock")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}
			if len(acct.Transfer) == 0 {
				orQuit(fmt.Errorf("Transfer key for %s not set", *name))
			}

			if *verbose {
				fmt.Printf(
					"Claiming node reward for %s\n",
					acct.Address.String(),
				)
			}

			tx := ndau.NewClaimNodeReward(
				acct.Address,
				sequence(conf, acct.Address),
				acct.TransferPrivateK(keys),
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), tx)
			finish(*verbose, resp, err, "claim-node-reward")
		}
	}
}
