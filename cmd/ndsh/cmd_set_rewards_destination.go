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
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/pkg/errors"
)

// SetRewardsDestination sets an account's destination for EAI and node rewards
type SetRewardsDestination struct{}

var _ Command = (*SetRewardsDestination)(nil)

// Name implements Command
func (SetRewardsDestination) Name() string { return "set-rewards-destination set-rewards srd" }

type srtargs struct {
	Destination string `arg:"positional,required" help:"rewards destination"`
	Account     string `arg:"positional" help:"account whose rewards destination to set"`
	Update      bool   `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage       bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (srtargs) Description() string {
	return strings.TrimSpace(`
Set the rewards destination of an account.
	`)
}

// Run implements Command
func (SetRewardsDestination) Run(argvs []string, sh *Shell) (err error) {
	args := srtargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var target *address.Address
	target, _, err = sh.AddressOf(args.Destination)
	if err != nil {
		return errors.Wrap(err, "get address of rewards target")
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

	sh.VWrite("setting rewards target of %s to %s", acct.Address, target)

	tx := ndau.NewSetRewardsDestination(
		acct.Address,
		*target,
		acct.Data.Sequence+1,
		acct.PrivateValidationKeys...,
	)
	return sh.Dispatch(args.Stage, tx, acct, nil)
}
