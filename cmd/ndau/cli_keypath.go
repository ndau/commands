package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import cli "github.com/jawher/mow.cli"

func getKeypathSpec(optional bool) string {
	if optional {
		return "[-P=<keypath>]"
	}
	return "KEYPATH"
}

func getKeypathClosure(cmd *cli.Cmd, optional bool) func() string {
	var keypath *string

	if optional {
		keypath = cmd.StringOpt("P keypath", "", "derivation path of key")
	} else {
		keypath = cmd.StringArg("KEYPATH", "", "derivation path of key")
	}

	return func() string {
		if keypath == nil {
			return ""
		}
		return *keypath
	}
}
