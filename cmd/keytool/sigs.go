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
	"errors"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

func cmdSign(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"%s %s [-r]",
		getKeySpec("PVT"),
		getDataSpec(false),
	)

	var (
		raw = cmd.BoolOpt("r raw", false, "emit the signature as raw bytes instead of ndau format")
	)

	getKey := getKeyClosure(cmd, "PVT", "sign with this private key")
	getData := getDataClosure(cmd, false)

	cmd.Action = func() {
		key, ok := getKey().(*signature.PrivateKey)
		if !ok {
			check(errors.New("signing requires a private key"))
		}
		data := getData()

		sig := key.Sign(data)
		if *raw {
			_, err := os.Stdout.Write(sig.Bytes())
			check(err)
		} else {
			sigb, err := sig.MarshalText()
			check(err)
			fmt.Println(string(sigb))
		}
	}
}

func cmdVerify(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"[-v] %s SIGNATURE %s",
		getKeySpec("PUB"),
		getDataSpec(false),
	)

	getKey := getKeyClosure(cmd, "PUB", "verify with this public key")
	getData := getDataClosure(cmd, false)

	verbose := cmd.BoolOpt("v verbose", false, "indicate success or failure on stdout in addition to the return code")
	sigi := cmd.StringArg("SIGNATURE", "", "verify this signature")

	cmd.Action = func() {
		key, ok := getKey().(*signature.PublicKey)
		if !ok {
			check(errors.New("verification requires a public key"))
		}
		data := getData()

		if sigi == nil || len(*sigi) == 0 {
			check(errors.New("signature not specified"))
		}
		var sig signature.Signature
		err := sig.UnmarshalText([]byte(*sigi))
		check(err)

		v := false
		if verbose != nil && *verbose {
			v = true
		}

		if sig.Verify(data, *key) {
			if v {
				fmt.Println("OK")
			}
		} else {
			if v {
				fmt.Println("NO MATCH")
			}
			cli.Exit(2)
		}
	}
}
