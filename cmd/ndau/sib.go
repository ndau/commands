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
	"github.com/ndau/ndau/pkg/tool"
)

func getSIB(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			info, resp, err := tool.GetSIB(tmnode(config.Node, nil, nil))
			if err == nil {
				var jsi []byte
				jsi, err = json.MarshalIndent(info, "", "  ")
				fmt.Println(string(jsi))
			}

			finish(*verbose, resp, err, "sib")
		}
	}
}
