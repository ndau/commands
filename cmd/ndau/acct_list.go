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
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getAccountAddr(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = getAddressSpec("")
		getAddress := getAddressClosure(cmd, "")
		cmd.Action = func() {
			fmt.Println(getAddress().String())
		}
	}
}

func getAccountList(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command("remote", "list accounts known to the ndau chain", getAccountListRemote(verbose))

		cmd.Action = func() {
			config := getConfig()
			config.EmitAccounts(os.Stdout)
		}
	}
}

func getAccountListRemote(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			accts, err := tool.GetAccountListBatch(tmnode(config.Node, nil, nil))
			orQuit(err)

			for _, acct := range accts {
				fmt.Println(acct)
			}
			finish(*verbose, nil, err, "account list remote")
		}
	}
}
