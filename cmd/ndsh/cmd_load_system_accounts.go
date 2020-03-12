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
	"reflect"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/ndau/ndau/pkg/tool"
	"github.com/ndau/ndaumath/pkg/address"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// LoadSystemAccounts loads the accounts described in system_accts.toml
type LoadSystemAccounts struct{}

var _ Command = (*LoadSystemAccounts)(nil)

// Name implements Command
func (LoadSystemAccounts) Name() string { return "load-system-accounts loadsa" }

type loadsaargs struct {
	Path  string `arg:"positional" help:"path to system_accts.toml"`
	Check bool   `help:"check that the system accounts loaded correspond with the active net"`
}

func (loadsaargs) Description() string {
	return strings.TrimSpace(`
Load system accounts from system_accts.toml
	`)
}

// Run implements Command
func (LoadSystemAccounts) Run(argvs []string, sh *Shell) (err error) {
	args := loadsaargs{
		Path: "~/.localnet/genesis_files/system_accounts.toml",
	}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	err = sh.LoadSystemAccts(args.Path)
	if err != nil {
		return
	}

	if args.Check {
		svs := []string{
			sv.CommandValidatorChangeAddressName,
			sv.NominateNodeRewardAddressName,
			sv.RecordPriceAddressName,
			sv.ReleaseFromEndowmentAddressName,
			sv.SetSysvarAddressName,
		}

		expect := make(map[string]address.Address)
		for _, name := range svs {
			ea, err := sh.SAAddr(name)
			if err != nil {
				return err
			}
			expect[name] = *ea
		}

		gotm, _, err := tool.Sysvars(sh.Node, svs...)
		if err != nil {
			return err
		}

		got := make(map[string]address.Address)
		for name, msgpdata := range gotm {
			var addr address.Address
			_, err = addr.UnmarshalMsg(msgpdata)
			if err != nil {
				return errors.Wrap(err, name)
			}
			got[name] = addr
		}

		if !reflect.DeepEqual(expect, got) {
			sh.systemAccts = nil
			return errors.New("blockchain sysvar data did not match configured; discarding configuration")
		}
	}

	return
}
