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
	"encoding/base64"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndau/pkg/tool"
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
			if len(acct.Validation) == 0 {
				orQuit(fmt.Errorf("Validation key for %s not set", *name))
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
				acct.ValidationPrivateK(*keys)...,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), tx)
			finish(*verbose, resp, err, "notify")
		}
	}
}
