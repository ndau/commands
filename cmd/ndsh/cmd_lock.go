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
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Lock locks an account
type Lock struct{}

var _ Command = (*Lock)(nil)

// Name implements Command
func (Lock) Name() string { return "lock" }

type lockargs struct {
	Period math.Duration `arg:"positional,required" help:"period of lock"`
	Target string        `arg:"positional" help:"account to lock"`
	Update bool          `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Notify bool          `arg:"-n,--notify" help:"notify the locked account immediately"`
	Stage  bool          `arg:"-S" help:"stage this tx; do not send it"`
}

func (lockargs) Description() string {
	return strings.TrimSpace(`
Lock an account.

With --notify, immediately notify the locked account.
	`)
}

// Run implements Command
func (Lock) Run(argvs []string, sh *Shell) (err error) {
	args := lockargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
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

	sh.VWrite("locking %s for %s", target.Address, args.Period)

	var tx metatx.Transactable
	tx = ndau.NewLock(
		target.Address,
		args.Period,
		target.Data.Sequence+1,
		target.PrivateValidationKeys...,
	)

	err = sh.Dispatch(args.Stage, tx, target, nil)
	if err != nil {
		return errors.Wrap(err, "dispatching lock tx")
	}

	if args.Notify {
		sh.VWrite("notifying...")
		tx = ndau.NewNotify(
			target.Address,
			target.Data.Sequence+1,
			target.PrivateValidationKeys...,
		)
		err = sh.Dispatch(args.Stage, tx, target, nil)
		if err != nil {
			return errors.Wrap(err, "dispatching notify tx")
		}
	}
	return
}
