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

func getDelegates(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			conf := getConfig()
			delegates, resp, err := tool.GetDelegates(tmnode(conf.Node, nil, nil))

			// errors get logged in the Info field
			fmt.Fprint(os.Stderr, resp.Response.GetInfo())

			// display the delegates
			for node, delegated := range delegates {
				fmt.Printf("%s:\n", node)
				for _, d := range delegated {
					fmt.Printf("  %s\n", d)
				}
			}

			finish(*verbose, resp, err, "show-delegates")
		}
	}
}
