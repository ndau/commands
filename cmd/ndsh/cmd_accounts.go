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
	"github.com/alexflint/go-arg"
)

// ListAccounts lists all known accounts
type ListAccounts struct{}

var _ Command = (*ListAccounts)(nil)

// Name implements Command
func (ListAccounts) Name() string { return "accounts list" }

// Run implements Command
func (ListAccounts) Run(argvs []string, sh *Shell) (err error) {
	args := struct {
	}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	for acct, nicknames := range sh.Accts.Reverse() {
		acct.display(sh, nicknames)
	}
	return
}
