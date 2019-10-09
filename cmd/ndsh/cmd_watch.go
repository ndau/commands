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
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

// Watch adds foreign accounts, sometimes with nicknames
type Watch struct{}

var _ Command = (*Watch)(nil)

// Name implements Command
func (Watch) Name() string { return "watch" }

type watchargs struct {
	Address   address.Address `arg:"positional,required" help:"watch this account"`
	Nicknames []string        `arg:"-n,separate" help:"short nicknames which can refer to this account."`
	NewOK     bool            `arg:"-N,--new-ok" help:"don't complain if account does not yet exist on blockchain"`
}

func (watchargs) Description() string {
	return strings.TrimSpace(`
Watch accounts for which you do not possess the private keys.

This is most useful to declare short nicknames for destination accounts.
	`)
}

// Run implements Command
func (Watch) Run(argvs []string, sh *Shell) (err error) {
	args := watchargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	a := &Account{
		Address: args.Address,
	}
	err = a.Update(sh, sh.Write)
	if args.NewOK && IsAccountDoesNotExist(err) {
		err = nil
	}
	if err != nil {
		return
	}

	sh.Accts.Add(a, args.Nicknames...)

	return
}
