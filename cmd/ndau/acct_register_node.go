package main

import (
	"encoding/base64"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getRegisterNode(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME DISTRIBUTION_SCRIPT"

		var (
			name       = cmd.StringArg("NAME", "", "Name of node to register")
			distScript = cmd.StringArg(
				"DISTRIBUTION_SCRIPT",
				"",
				"non-padded base64 encoding of chaincode distribution script",
			)
		)

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}
			if len(acct.Transfer) == 0 {
				orQuit(fmt.Errorf("Transfer key for %s not set", *name))
			}

			script, err := base64.RawStdEncoding.DecodeString(*distScript)
			orQuit(err)

			if *verbose {
				fmt.Printf(
					"Registering node %s\n",
					acct.Address,
				)
			}

			tx := ndau.NewRegisterNode(
				acct.Address, script, acct.Ownership.Public,
				sequence(conf, acct.Address),
				acct.TransferPrivateK(*keys)...,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), tx)
			finish(*verbose, resp, err, "notify")
		}
	}
}
