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
	"github.com/pkg/errors"
)

// ClaimNodeReward claims node rewards for which an account has been nominated
type ClaimNodeReward struct{}

var _ Command = (*ClaimNodeReward)(nil)

// Name implements Command
func (ClaimNodeReward) Name() string { return "claim-node-reward cnr" }

type cnrargs struct {
	Account string `arg:"positional" help:"account to claim node reward for"`
	Update  bool   `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage   bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (cnrargs) Description() string {
	return strings.TrimSpace(`
Claim a node reward.
	`)
}

// Run implements Command
func (ClaimNodeReward) Run(argvs []string, sh *Shell) (err error) {
	args := cnrargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
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

	sh.VWrite("claiming node reward for %s", acct.Address)

	tx := ndau.NewClaimNodeReward(
		acct.Address,
		acct.Data.Sequence+1,
		acct.PrivateValidationKeys...,
	)
	return sh.Dispatch(args.Stage, tx, acct, nil)
}
