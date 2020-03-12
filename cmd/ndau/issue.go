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
	config "github.com/ndau/ndau/pkg/tool.config"
	"github.com/pkg/errors"
)

func getIssue(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = getNdauSpec()

		getNdau := getNdauClosure(cmd)

		cmd.Action = func() {
			ndauQty := getNdau()

			if *verbose {
				fmt.Printf("Issue: %s ndau\n", ndauQty)
			}

			conf := getConfig()
			if conf.RFE == nil {
				orQuit(errors.New("RFE data (for issuance) not set in tool config"))
			}

			keys := config.FilterK(conf.RFE.Keys, *keys)

			issue := ndau.NewIssue(
				ndauQty,
				sequence(conf, conf.RFE.Address),
				keys...,
			)

			result, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), issue)
			finish(*verbose, result, err, "issue")
		}
	}
}
