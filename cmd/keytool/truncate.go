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
)

func cmdTruncate(cmd *cli.Cmd) {
	cmd.Spec = getKeySpec("")

	getKey := getKeyClosure(cmd, "", "key to truncate")

	cmd.Action = func() {
		key := getKey()
		key.Truncate()
		keyB, err := key.MarshalText()
		check(err)
		fmt.Println(string(keyB))
	}
}
