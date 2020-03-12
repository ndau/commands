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
	"github.com/pkg/errors"
)

// Notify an account
type Notify struct{}

var _ Command = (*Notify)(nil)

// Name implements Command
func (Notify) Name() string { return "notify" }

type notifyargs struct {
	Target string `arg:"positional" help:"account to notify"`
	Update bool   `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage  bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (notifyargs) Description() string {
	return strings.TrimSpace(`
Notify an account.
	`)
}

// Run implements Command
func (Notify) Run(argvs []string, sh *Shell) (err error) {
	args := notifyargs{}

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

	sh.VWrite("notifying %s", target.Address)

	tx := ndau.NewNotify(
		target.Address,
		target.Data.Sequence+1,
		target.PrivateValidationKeys...,
	)
	return sh.Dispatch(args.Stage, tx, target, nil)
}
