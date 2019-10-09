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
	"errors"
	"strings"

	"github.com/alexflint/go-arg"
)

// Exit leaves the shell
type Exit struct{}

var _ Command = (*Exit)(nil)

// Name implements Command
func (Exit) Name() string { return "exit quit" }

// Run implements Command
func (Exit) Run(argvs []string, sh *Shell) (err error) {
	args := struct {
		Error []string `arg:"positional" help:"Error message to pass out to the outer context"`
	}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	if len(args.Error) > 0 {
		err = errors.New(strings.Join(args.Error, " "))
	}
	sh.Exit(err)
	return
}
