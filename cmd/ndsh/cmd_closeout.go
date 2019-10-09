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

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/sirupsen/logrus"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/pkg/errors"
)

// Closeout an account
type Closeout struct{}

var _ Command = (*Closeout)(nil)

// Name implements Command
func (Closeout) Name() string { return "closeout" }

type closeoutargs struct {
	Into       string `arg:"positional,required" help:"account to move existing funds into"`
	Account    string `arg:"positional" help:"account to close"`
	Update     bool   `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Iterations uint   `arg:"-i" help:"max iterations to attempt before giving up"`
	Stage      bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (closeoutargs) Description() string {
	return strings.TrimSpace(`
Close out an account into another.

This simply moves all available funds, minus fees and SIB, into the new account
in a single Tx, maximizing efficiency.

Note that because fees and SIB are computed dynamically and the nature of the
computation may change in real time, there is no general-purpose formula for
determining the correct amount to send. Instead, this simply iterates a number
of times, attempting to maximize the balance transferred.
	`)
}

// Run implements Command
func (Closeout) Run(argvs []string, sh *Shell) (err error) {
	args := closeoutargs{
		Iterations: 32,
	}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var into *address.Address
	into, _, err = sh.AddressOf(args.Into)
	if err != nil {
		return errors.Wrap(err, "into")
	}

	var acct *Account
	acct, err = sh.Accts.Get(args.Account)
	if err != nil {
		return errors.Wrap(err, "account")
	}

	if acct.Data == nil || args.Update {
		err = acct.Update(sh, sh.Write)
		if IsAccountDoesNotExist(err) {
			err = nil
		}
		if err != nil {
			return
		}
	}

	avail, err := acct.Data.AvailableBalance()
	if err != nil {
		return errors.Wrap(err, "computing available balance")
	}
	if avail != acct.Data.Balance {
		sh.Write("WARN: full balance (%s ndau) != available balance (%s ndau)", acct.Data.Balance, avail)
	}
	qty := avail

	var tx metatx.Transactable
	prev := math.Ndau(0)
	for i := uint(0); i < args.Iterations && qty != prev && qty > 0; i++ {
		sh.VWrite("%2d: %s ndau", i+1, qty)
		tx = ndau.NewTransfer(
			acct.Address,
			*into,
			qty,
			acct.Data.Sequence+1,
			acct.PrivateValidationKeys...,
		)
		logger := logrus.New()
		fee, sib, _, err := tool.Prevalidate(sh.Node, tx, logger)
		if fee == 0 && sib == 0 && err != nil {
			return errors.Wrap(err, "prevalidating")
		}
		// if the error is not nil but fee and sib are unset, assume the error
		// is that we can't afford the tx, and ignore it
		prev = qty
		if qty+fee+sib > avail {
			qty = avail - (fee + sib)
		} else if qty+fee+sib < avail {
			// get halfway back to the available balance
			dist := avail - (qty + fee + sib)
			qty += dist / 2
		} else {
			// we were dead on; did it work?
			if err == nil {
				break
			}
			qty--
		}
	}

	sh.VWrite("closing out %s into %s by moving %s ndau", acct.Address, *into, qty)
	return sh.Dispatch(args.Stage, tx, acct, nil)
}
