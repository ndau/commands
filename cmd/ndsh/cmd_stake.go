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

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/pkg/errors"
)

// Stake an account
type Stake struct{}

var _ Command = (*Stake)(nil)

// Name implements Command
func (Stake) Name() string { return "stake" }

type stakeargs struct {
	Rules   string `arg:"positional,required" help:"rules account"`
	StakeTo string `arg:"positional,required" help:"account staked to"`
	Target  string `arg:"positional" help:"account to stake"`
	Qty     int64  `arg:"-q" help:"amount to stake. If not set, use full balance"`
	Update  bool   `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage   bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (stakeargs) Description() string {
	return strings.TrimSpace(`
Stake an account.
	`)
}

// Run implements Command
func (Stake) Run(argvs []string, sh *Shell) (err error) {
	args := stakeargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var rules, stakeTo *address.Address
	rules, _, err = sh.AddressOf(args.Rules)
	if err != nil {
		return errors.Wrap(err, "get address of Rules")
	}
	stakeTo, _, err = sh.AddressOf(args.StakeTo)
	if err != nil {
		return errors.Wrap(err, "get address of StakeTo")
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

	var qty math.Ndau
	if args.Qty == 0 {
		qty = target.Data.Balance
	} else {
		qty = math.Ndau(args.Qty)
	}

	sh.VWrite("staking %s ndau from %s to %s using rules %s", qty, target.Address, stakeTo, rules)

	tx := ndau.NewStake(
		target.Address,
		*rules,
		*stakeTo,
		qty,
		target.Data.Sequence+1,
		target.PrivateValidationKeys...,
	)
	return sh.Dispatch(args.Stage, tx, target, nil)
}
