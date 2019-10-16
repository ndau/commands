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
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/pkg/errors"
)

func getRfe(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s",
			getNdauSpec(),
			getAddressSpec(""),
		)

		getNdau := getNdauClosure(cmd)
		getAddress := getAddressClosure(cmd, "")

		cmd.Action = func() {
			ndauQty := getNdau()
			address := getAddress()

			if *verbose {
				fmt.Printf("Release from endowment: %s ndau to %s\n", ndauQty, address)
			}

			conf := getConfig()
			if conf.RFE == nil {
				orQuit(errors.New("RFE data not set in tool config"))
			}

			keys := config.FilterK(conf.RFE.Keys, *keys)

			rfe := ndau.NewReleaseFromEndowment(
				address,
				ndauQty,
				sequence(conf, conf.RFE.Address),
				keys...,
			)

			result, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), rfe)
			finish(*verbose, result, err, "rfe")
		}
	}
}
