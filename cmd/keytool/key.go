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
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/ndau/ndaumath/pkg/key"
	"github.com/ndau/ndaumath/pkg/signature"
)

// ktype should always be "", "PUB", or "PVT"
func keytype(ktype string) string {
	return strings.ToUpper(strings.TrimSpace(ktype)) + "KEY"
}

func getKeySpec(ktype string) string {
	return fmt.Sprintf("(%s | --stdin)", keytype(ktype))
}

func getKeyClosure(cmd *cli.Cmd, ktype string, desc string) func() signature.Key {
	key := cmd.StringArg(keytype(ktype), "", desc)
	stdin := cmd.BoolOpt("S stdin", false, "if set, read the key from stdin")

	return func() signature.Key {
		var keys string
		if stdin != nil && *stdin {
			in := bufio.NewScanner(os.Stdin)
			if !in.Scan() {
				check(errors.New("stdin selected but empty"))
			}
			check(in.Err())
			keys = in.Text()
		} else if key != nil && len(*key) > 0 {
			keys = *key
		} else {
			check(errors.New("no or multiple keys input--this should be unreachable"))
		}

		k, err := signature.ParseKey(keys)
		check(err)
		return k
	}
}

func getKeyClosureHD(cmd *cli.Cmd, ktype string, desc string) func() *key.ExtendedKey {
	getKey := getKeyClosure(cmd, ktype, desc)
	return func() *key.ExtendedKey {
		k := getKey()
		ek, err := key.FromSignatureKey(k)
		check(err)
		return ek
	}
}
