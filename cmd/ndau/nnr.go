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
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strconv"

	cli "github.com/jawher/mow.cli"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndau/pkg/tool"
	config "github.com/ndau/ndau/pkg/tool.config"
	"github.com/pkg/errors"
)

func getNNR(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "(RANDOM | -g)"

		generate := cmd.BoolOpt("g generate", false, "generate a random number using the best crypto RNG available")
		randS := cmd.StringArg("RANDOM", "", "specify a random number in the u64 range")

		cmd.Action = func() {
			random := int64(0)
			if generate != nil && *generate {
				bytes := make([]byte, 8)
				_, err := rand.Read(bytes)
				orQuit(err)
				random = int64(binary.BigEndian.Uint64(bytes))
			} else if randS != nil && *randS != "" {
				var err error
				random, err = strconv.ParseInt(*randS, 0, 64)
				orQuit(err)
			} else {
				orQuit(errors.New("no random number specified"))
			}

			if *verbose {
				fmt.Printf("Official Random Number: %d\n", random)
			}

			conf := getConfig()
			if conf.NNR == nil {
				orQuit(errors.New("NNR data not set in tool config"))
			}

			keys := config.FilterK(conf.NNR.Keys, *keys)

			nnr := ndau.NewNominateNodeReward(
				random,
				sequence(conf, conf.NNR.Address),
				keys...,
			)

			result, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), nnr)
			finish(*verbose, result, err, "nnr")
		}
	}
}
