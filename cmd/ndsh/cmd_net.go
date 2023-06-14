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
	tmclient "github.com/oneiro-ndev/tendermint.0.32.3/rpc/client"
)

// Net prints the net currently connected to, or updates it
type Net struct{}

var _ Command = (*Net)(nil)

// Name implements Command
func (Net) Name() string { return "net" }

// Run implements Command
func (Net) Run(argvs []string, sh *Shell) (err error) {
	args := struct {
		Set string `help:"switch networks to this network. WARNING: this can cause inconsistent state, only do this if you know what you're doing."`
		Num int    `help:"node number to use when switching networks"`
	}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	if args.Set != "" {
		var client tmclient.ABCIClient
		client, err = getClient(args.Set, args.Num)
		if err != nil {
			return
		}
		sh.Node = client
	}
	// ClientURL gets updated as a side-effect of getClient
	// so does RecoveryURL
	sh.Write("    node: %s\nrecovery: %s", ClientURL, RecoveryURL)
	return
}
