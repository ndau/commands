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
	"fmt"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
)

func nameFor(id string) string {
	if id == "" {
		return "NAME"
	}
	return fmt.Sprintf("%s_NAME", id)
}

func argFor(id string) string {
	if id == "" {
		return "a address"
	}
	return strings.ToLower(fmt.Sprintf("%s-address", id))
}

func getAddressSpec(id string) string {
	if id == "" {
		return "(NAME | -a=<ADDRESS>)"
	}
	return fmt.Sprintf(
		"(%s | --%s=<%s>)",
		nameFor(id), argFor(id), strings.ToUpper(argFor(id)),
	)
}

func getAddressClosure(cmd *cli.Cmd, id string) func() address.Address {
	name := cmd.StringArg(nameFor(id), "", fmt.Sprintf("Name of %s account", id))
	addr := cmd.StringOpt(argFor(id), "", fmt.Sprintf("%s Address", id))

	return func() address.Address {
		if addr != nil && len(*addr) > 0 {
			a, err := address.Validate(*addr)
			orQuit(err)
			return a
		}
		if name != nil && len(*name) > 0 {
			config := getConfig()
			acct, hasAcct := config.Accounts[*name]
			if hasAcct {
				return acct.Address
			}
			orQuit(errors.New("No such named account"))
		}
		orQuit(errors.New("Neither name nor address specified"))
		return address.Address{}
	}
}
