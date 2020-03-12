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
	"github.com/pkg/errors"
)

func getRecordPrice(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s",
			getNanocentSpec(),
		)

		getNanocent := getNanocentClosure(cmd)

		cmd.Action = func() {
			nanocentQty := getNanocent()

			if *verbose {
				fmt.Printf(
					"RecordPrice %d nanocents\n",
					nanocentQty,
				)
			}

			conf := getConfig()
			if conf.RecordPrice == nil {
				orQuit(errors.New("RecordPrice keys not set in config"))
			}

			// construct the recordPrice
			recordPrice := ndau.NewRecordPrice(
				nanocentQty,
				sequence(conf, conf.RecordPrice.Address),
				conf.RecordPrice.Keys...,
			)

			tresp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), recordPrice)
			finish(*verbose, tresp, err, "record-price")
		}
	}
}
