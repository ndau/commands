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

func getAccountChangeRecourse(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"NAME %s",
			getDurationSpec(),
		)

		var name = cmd.StringArg("NAME", "", "Name of account of which to change recourse period")
		getDuration := getDurationClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			duration := getDuration()

			ad, hasAd := config.Accounts[*name]
			if !hasAd {
				orQuit(errors.New("No such account found"))
			}
			if ad.Validation == nil {
				orQuit(errors.New("Address validation key not set"))
			}

			cep := ndau.NewChangeRecoursePeriod(
				ad.Address,
				duration,
				sequence(config, ad.Address),
				ad.ValidationPrivateK(*keys)...,
			)

			if *verbose {
				fmt.Printf(
					"Change Recourse Period for %s (%s) to %s\n",
					*name,
					ad.Address,
					duration,
				)
			}

			resp, err := tool.SendCommit(tmnode(config.Node, emitJSON, compact), cep)
			finish(*verbose, resp, err, "change-recourse-period")
		}
	}
}
