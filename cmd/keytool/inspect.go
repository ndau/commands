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
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

func cmdInspect(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"(%s | --sig=<SIGNATURE>)",
		getKeySpec(""),
	)

	getKey := getKeyClosure(cmd, "", "key to be inspected")
	var (
		sigp = cmd.StringOpt("s sig", "", "signature to inspect")
	)

	cmd.Action = func() {
		if sigp != nil && *sigp != "" {
			sig, err := signature.ParseSignature(*sigp)
			check(err)

			fmt.Printf("%10s: %s\n", "algorithm", signature.NameOf(sig.Algorithm()))
			fmt.Printf("%10s: %x\n", "data", sig.Bytes())
			return
		}

		key := getKey()
		ktype := "public"
		if signature.IsPrivate(key) {
			ktype = "private"
		}

		fmt.Printf("%10s: %s\n", "type", ktype)
		fmt.Printf("%10s: %s\n", "algorithm", signature.NameOf(key.Algorithm()))
		fmt.Printf("%10s: %x\n", "key", key.KeyBytes())

		extraBytes := key.ExtraBytes()

		if len(extraBytes) != 0 {
			fmt.Printf("\n%20s: %x\n", "extended key depth", extraBytes[0])
			fmt.Printf("%20s: %x\n", "parent fingerprint", extraBytes[1:4])
			fmt.Printf("%20s: %x\n", "child number", extraBytes[4:8])
			fmt.Printf("%20s: %x\n", "chain code", extraBytes[8:40])
		}
	}
}
