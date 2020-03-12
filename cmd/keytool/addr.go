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

	cli "github.com/jawher/mow.cli"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
)

func getKindSpec() string {
	// mow.cli ensures with this that only one option is specified
	return "[-k=<kind> | -a | -n | -e | -x | -b | -m]"
}

func getKindClosure(cmd *cli.Cmd) func() byte {
	var (
		pkind  = cmd.StringOpt("k kind", string(address.KindUser), "manually specify address kind")
		kuser  = cmd.BoolOpt("a user", false, "address kind: user (default)")
		kndau  = cmd.BoolOpt("n ndau", false, "address kind: ndau")
		kendow = cmd.BoolOpt("e endowment", false, "address kind: endowment")
		kxchng = cmd.BoolOpt("x exchange", false, "address kind: exchange")
		kbpc   = cmd.BoolOpt("b bpc", false, "address kind: bpc")
		kmm    = cmd.BoolOpt("m market-maker", false, "address kind: market maker")
	)

	return func() byte {
		var kind byte
		switch {
		case kuser != nil && *kuser:
			kind = address.KindUser
		case kndau != nil && *kndau:
			kind = address.KindNdau
		case kendow != nil && *kendow:
			kind = address.KindEndowment
		case kxchng != nil && *kxchng:
			kind = address.KindExchange
		case kbpc != nil && *kbpc:
			kind = address.KindBPC
		case kmm != nil && *kmm:
			kind = address.KindMarketMaker
		default:
			kindString := *pkind // never nil dereference; defaults to user
			if len(kindString) != 1 {
				check(fmt.Errorf("invalid kind length: '%s'", kindString))
			}
			kind = kindString[0]
		}

		if !address.IsValidKind(kind) {
			check(fmt.Errorf("invalid kind byte: %x", kind))
		}
		return kind
	}
}

func cmdAddr(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"%s %s",
		getKeySpec("PUB"),
		getKindSpec(),
	)

	getKey := getKeyClosure(cmd, "PUB", "get address from this key")
	getKind := getKindClosure(cmd)

	cmd.Action = func() {
		key := getKey()
		_, ok := key.(*signature.PublicKey)
		if !ok {
			check(errors.New("addresses can only be generated from public keys"))
		}

		kind := getKind()

		addr, err := address.Generate(kind, key.KeyBytes())
		check(err)
		fmt.Println(addr)
	}
}
