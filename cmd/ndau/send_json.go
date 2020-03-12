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
	"github.com/ndau/ndau/pkg/tool"
)

func getSendJSON(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s",
			getJSONTXSpec(),
		)

		getJSONTX := getJSONTXClosure(cmd)

		cmd.Action = func() {
			tx := getJSONTX()
			conf := getConfig()
			resp, err := tool.SendCommit(tmnode(conf.Node, nil, nil), tx)
			finish(*verbose, resp, err, "send-json")
		}
	}
}
