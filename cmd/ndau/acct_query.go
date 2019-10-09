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
	"encoding/json"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getAccountQuery(verbose, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf("%s", getAddressSpec(""))
		getAddress := getAddressClosure(cmd, "")

		cmd.Action = func() {
			address := getAddress()
			config := getConfig()
			ad, resp, err := tool.GetAccount(tmnode(config.Node, emitJSON, compact), address)
			if err != nil {
				finish(*verbose, resp, err, "account")
			}
			jsb, err := json.MarshalIndent(ad, "", "  ")
			fmt.Println(string(jsb))
			finish(*verbose, resp, err, "account")
		}
	}
}
