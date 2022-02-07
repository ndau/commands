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

	"github.com/ndau/ndaumath/pkg/address"
	math "github.com/ndau/ndaumath/pkg/types"

	"github.com/alexflint/go-arg"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/pkg/errors"
)

// Unstake an account
type Unstake struct{}

var _ Command = (*Unstake)(nil)

// Name implements Command
func (Unstake) Name() string { return "unstake" }

type unstakeargs struct {
	Rules       string `arg:"positional,required" help:"rules account"`
	UnstakeFrom string `arg:"positional,required" help:"account staked to"`
	Target      string `arg:"positional" help:"account to unstake"`
	Qty         int64  `arg:"positional" help:"amount to unstake"`
	Update      bool   `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage       bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (unstakeargs) Description() string {
	return strings.TrimSpace(`
Unstake an account.
	`)
}

// Run implements Command
func (Unstake) Run(argvs []string, sh *Shell) (err error) {
	args := unstakeargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var rules, unstakeFrom *address.Address
	rules, _, err = sh.AddressOf(args.Rules)
	if err != nil {
		return errors.Wrap(err, "get address of Rules")
	}
	unstakeFrom, _, err = sh.AddressOf(args.UnstakeFrom)
	if err != nil {
		return errors.Wrap(err, "get address of UnstakeFrom")
	}

	var target *Account
	target, err = sh.Accts.Get(args.Target)
	if err != nil {
		return errors.Wrap(err, "account")
	}

	if target.Data == nil || args.Update {
		err = target.Update(sh, sh.Write)
		if IsAccountDoesNotExist(err) {
			err = nil
		}
		if err != nil {
			return
		}
	}

	// TODO: Add error checking here

	var qty math.Ndau
	qty = math.Ndau(args.Qty)

	sh.VWrite("unstaking %s ndau from %s's stake to %s using rules %s", qty, target.Address, unstakeFrom, rules)

	tx := ndau.NewUnstake(
		target.Address,
		*rules,
		*unstakeFrom,
		qty,
		target.Data.Sequence+1,
		target.PrivateValidationKeys...,
	)
	return sh.Dispatch(args.Stage, tx, target, nil)
}
