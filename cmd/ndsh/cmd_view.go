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
	"encoding/json"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/pkg/errors"
	"github.com/savaki/jq"
)

// View views an account
type View struct{}

var _ Command = (*View)(nil)

// Name implements Command
func (View) Name() string { return "view show" }

type viewargs struct {
	Account string `arg:"positional" help:"view this account"`
	Update  bool   `arg:"-u" help:"update this account from the blockchain before viewing"`
	Struct  bool   `help:"show the whole account struct, not just the account data"`
	JQ      string `help:"filter output json by this jq expression"`
}

func (viewargs) Description() string {
	return strings.TrimSpace(`
View an account's data.

By default, this operates only on cached data. To get current data from the
blockchain, use the --update flag.

Note that the JQ implementation used here is a pure-go reimplementation, not
bindings to libjq. This is convenient for compilation, but it means that the
only features actually implemented are simple selectors.
	`)
}

// Run implements Command
func (View) Run(argvs []string, sh *Shell) (err error) {
	args := viewargs{}

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
		return
	}

	acct.display(sh, sh.Accts.Reverse()[acct])

	if args.Update {
		if sh.Verbose {
			sh.Write("communicating with blockchain...")
		}
		err = acct.Update(sh, sh.Write)
		if err != nil {
			return
		}
	}

	var data interface{}
	if args.Struct {
		data = acct
	} else {
		data = acct.Data
	}

	jsdata, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		err = errors.Wrap(err, "marshalling account data to json")
		return
	}

	if args.JQ != "" {
		op, err := jq.Parse(args.JQ)
		if err != nil {
			return errors.Wrap(err, "parsing JQ selector")
		}
		jsdata, err = op.Apply(jsdata)
		if err != nil {
			return errors.Wrap(err, "applying JQ selector")
		}
	}

	sh.Write(string(jsdata))
	return
}
