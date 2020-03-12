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
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndau/pkg/tool"
)

func getTransfer(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s",
			getNdauSpec(),
			getAddressSpec("FROM"),
			getAddressSpec("TO"),
		)

		getNdau := getNdauClosure(cmd)
		getFrom := getAddressClosure(cmd, "FROM")
		getTo := getAddressClosure(cmd, "TO")

		cmd.Action = func() {
			ndauQty := getNdau()
			from := getFrom()
			to := getTo()

			if *verbose {
				fmt.Printf(
					"Transfer %s ndau from %s to %s\n",
					ndauQty, from, to,
				)
			}

			conf := getConfig()

			// ensure we know the private validation key of this account
			fromAcct, hasAcct := conf.Accounts[from.String()]
			if !hasAcct || fromAcct == nil {
				orQuit(fmt.Errorf("Account for address '%s' not found in config", from))
			}
			if fromAcct.Validation == nil {
				orQuit(fmt.Errorf("From acct validation key not set"))
			}

			// construct the transfer
			transfer := ndau.NewTransfer(
				from, to,
				ndauQty,
				sequence(conf, from),
				fromAcct.ValidationPrivateK(*keys)...,
			)

			tresp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), transfer)
			finish(*verbose, tresp, err, "transfer")
		}
	}
}
