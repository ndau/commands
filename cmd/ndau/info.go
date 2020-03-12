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
	"encoding/base64"
	"encoding/hex"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/ndau/ndau/pkg/tool"
	"github.com/tendermint/tendermint/rpc/client"
)

func getInfo(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		key := cmd.BoolOpt("k key", false, "when set, emit the public key of the connected node")
		emithex := cmd.BoolOpt("x hex", false, "when set, emit the key as hex instead of base64")
		pwr := cmd.BoolOpt("p power", false, "when set, emit the power of the connected node")

		cmd.Spec = "[-k [-x]] [-p]"

		cmd.Action = func() {
			config := getConfig()
			info, err := tool.Info(tmnode(config.Node, nil, nil).(*client.HTTP))

			if *key {
				b := info.ValidatorInfo.PubKey.Bytes()
				var p string
				if *emithex {
					p = hex.EncodeToString(b)
				} else {
					p = base64.RawStdEncoding.EncodeToString(b)
				}
				fmt.Println(p)
			}
			if *pwr {
				fmt.Println(info.ValidatorInfo.VotingPower)
			}
			if !(*key || *pwr) {
				vb := true
				verbose = &vb
			}
			finish(*verbose, info, err, "info")
		}
	}
}
