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
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndau/pkg/version"
)

// Version displays version information
type Version struct{}

var _ Command = (*Version)(nil)

// Name implements Command
func (Version) Name() string { return "version" }

// Run implements Command
func (Version) Run(argvs []string, sh *Shell) (err error) {
	args := struct{}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	local, err := version.Get()
	if err != nil {
		sh.Write("getting local version: %s", err)
	}
	remote, _, err := tool.Version(sh.Node)
	if err != nil {
		sh.Write("getting remote version: %s", err)
	}

	sh.Write(" local: %s\nremote: %s", local, remote)

	return
}
