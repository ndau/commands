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
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// TransferAndLock transfers ndau to a newly-created account and locks it
type TransferAndLock struct{}

var _ Command = (*TransferAndLock)(nil)

// Name implements Command
func (TransferAndLock) Name() string { return "transfer-lock tnl" }

type transferlockargs struct {
	Qty      string        `arg:"positional,required" help:"qty to transfer in ndau"`
	From     string        `arg:"positional,required" help:"account to transfer from. Use \"\" for inference"`
	To       string        `arg:"positional,required" help:"account to transfer to. Any full address is valid even if not otherwise known."`
	Duration math.Duration `arg:"positional,required" help:"Duration of the lock"`
	Stage    bool          `arg:"-S" help:"stage this tx; do not send it"`
}

func (transferlockargs) Description() string {
	return strings.TrimSpace(`
Transfer ndau from one account to another, locking the recipient.
	`)
}

// Run implements Command
func (TransferAndLock) Run(argvs []string, sh *Shell) (err error) {
	args := transferlockargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var from, to *Account
	from, err = sh.Accts.Get(args.From)
	if err != nil {
		return
	}

	var toaddr *address.Address
	toaddr, to, err = sh.AddressOf(args.To)
	if err != nil {
		return errors.Wrap(err, "to")
	}

	var qty math.Ndau
	qty, err = math.ParseNdau(args.Qty)
	if err != nil {
		return
	}

	sh.VWrite("transfering %s ndau (%d napu) from %s to %s and locking for %s", qty, qty, from.Address, toaddr, args.Duration)

	tx := ndau.NewTransferAndLock(
		from.Address,
		*toaddr,
		qty,
		args.Duration,
		from.Data.Sequence+1,
		from.PrivateValidationKeys...,
	)

	err = sh.Dispatch(args.Stage, tx, from, nil)
	if err != nil {
		return
	}

	if to != nil {
		err = to.Update(sh, sh.Write)
	}
	return
}
