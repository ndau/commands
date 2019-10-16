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
	"github.com/pkg/errors"
)

func getRecordEndowmentNAV(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
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
					"RecordEndowmentNAV %d nanocents\n",
					nanocentQty,
				)
			}

			conf := getConfig()
			if conf.RFE == nil {
				orQuit(errors.New("RFE keys for RecordEndowmentNAV not set in config"))
			}

			// construct the RecordEndowmentNAV
			RecordEndowmentNAV := ndau.NewRecordEndowmentNAV(
				nanocentQty,
				sequence(conf, conf.RFE.Address),
				conf.RFE.Keys...,
			)

			tresp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), RecordEndowmentNAV)
			finish(*verbose, tresp, err, "record-price")
		}
	}
}
