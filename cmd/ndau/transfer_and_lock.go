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

func getTransferAndLock(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s %s",
			getNdauSpec(),
			getAddressSpec("FROM"),
			getAddressSpec("TO"),
			getDurationSpec(),
		)

		getNdau := getNdauClosure(cmd)
		getFrom := getAddressClosure(cmd, "FROM")
		getTo := getAddressClosure(cmd, "TO")
		getDuration := getDurationClosure(cmd)

		cmd.Action = func() {
			ndauQty := getNdau()
			from := getFrom()
			to := getTo()
			duration := getDuration()

			if *verbose {
				fmt.Printf(
					"Transfer %s ndau from %s to %s, locking for %s\n",
					ndauQty, from, to, duration,
				)
			}

			conf := getConfig()

			// ensure we know the private validation key of this account
			fromAcct, hasAcct := conf.Accounts[from.String()]
			if !hasAcct {
				orQuit(fmt.Errorf("From account '%s' not found", fromAcct.Name))
			}
			if fromAcct.Validation == nil {
				orQuit(fmt.Errorf("From acct validation key not set"))
			}

			// construct the transfer
			transfer := ndau.NewTransferAndLock(
				from, to,
				ndauQty,
				duration,
				sequence(conf, from),
				fromAcct.ValidationPrivateK(*keys)...,
			)

			tresp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), transfer)
			finish(*verbose, tresp, err, "transferandlock")
		}
	}
}
