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

func getStake(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"NAME %s %s %s",
			getAddressSpec("RULES"),
			getAddressSpec("STAKETO"),
			getNdauSpec(),
		)

		var acctName = cmd.StringArg("NAME", "", "Name of account to stake")
		getRules := getAddressClosure(cmd, "RULES")
		getStakeTo := getAddressClosure(cmd, "STAKETO")
		getNdau := getNdauClosure(cmd)

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*acctName]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *acctName))
			}
			if len(acct.Validation) == 0 {
				orQuit(fmt.Errorf("Validation key for %s not set", *acctName))
			}

			rules := getRules()
			staketo := getStakeTo()
			qty := getNdau()

			if *verbose {
				fmt.Printf(
					"Staking %s ndau from acct %s to %s using rules %s\n",
					qty,
					acct.Address,
					staketo,
					rules,
				)
			}

			tx := ndau.NewStake(
				acct.Address,
				rules,
				staketo,
				qty,
				sequence(conf, acct.Address),
				acct.ValidationPrivateK(*keys)...,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), tx)
			finish(*verbose, resp, err, "notify")
		}
	}
}
