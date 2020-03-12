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
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndaumath/pkg/pricecurve"
	sv "github.com/ndau/system_vars/pkg/system_vars"
)

// RecordPrice records the current market price of ndau (in USD)
type RecordPrice struct{}

var _ Command = (*RecordPrice)(nil)

// Name implements Command
func (RecordPrice) Name() string { return "record-price" }

type recordpriceargs struct {
	Dollars string `arg:"positional,required" help:"record this quantity of dollars as the current price"`
	Stage   bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (recordpriceargs) Description() string {
	return strings.TrimSpace(`
Send a random number from which a node reward winner is derived.
	`)
}

// Run implements Command
func (RecordPrice) Run(argvs []string, sh *Shell) (err error) {
	args := recordpriceargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var magic *Account
	magic, err = sh.SAAcct(sv.RecordPriceAddressName, sv.RecordPriceValidationPrivateName)
	if err != nil {
		return
	}

	var nanocents pricecurve.Nanocent
	nanocents, err = pricecurve.ParseDollars(args.Dollars)
	if err != nil {
		return
	}

	sh.VWrite("recording current price: %d nanocents", nanocents)

	tx := ndau.NewRecordPrice(
		nanocents,
		magic.Data.Sequence+1,
		magic.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, nil, magic)
}
