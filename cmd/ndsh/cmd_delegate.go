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
	"github.com/pkg/errors"
)

// Delegate delegates to an account
type Delegate struct{}

var _ Command = (*Delegate)(nil)

// Name implements Command
func (Delegate) Name() string { return "delegate" }

type delegateargs struct {
	Target string `arg:"positional" help:"account to delegate"`
	Node   string `arg:"positional,required" help:"account delegated to"`
	Stage  bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (delegateargs) Description() string {
	return strings.TrimSpace(`
Delegate an account to another.

The delegation node becomes responsible for managing EAI calculations.
	`)
}

// Run implements Command
func (Delegate) Run(argvs []string, sh *Shell) (err error) {
	args := delegateargs{}

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

	var node *address.Address
	node, _, err = sh.AddressOf(args.Node)
	if err != nil {
		return errors.Wrap(err, "node")
	}

	sh.VWrite("delegating %s to %s", target.Address, node)

	tx := ndau.NewDelegate(
		target.Address,
		*node,
		target.Data.Sequence+1,
		target.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, target, nil)
}
