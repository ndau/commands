package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getAccountDelegate(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
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
			if acct.Validation == nil {
				orQuit(fmt.Errorf("Validation key for %s not set", *name))
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
				acct.ValidationPrivateK(*keys)...,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), tx)
			finish(*verbose, resp, err, "delegate")
		}
	}
}
