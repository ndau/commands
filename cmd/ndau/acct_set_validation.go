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
	cli "github.com/jawher/mow.cli"
	"github.com/ndau/metanode/pkg/meta/app/code"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndau/pkg/tool"
	config "github.com/ndau/ndau/pkg/tool.config"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/pkg/errors"
	rpc "github.com/oneiro-ndev/tendermint.0.32.3/rpc/core/types"
)

func getAccountSetValidation(verbose, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME"

		var name = cmd.StringArg("NAME", "", "Name of account")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			newKeys, err := acct.MakeValidationKey(nil)
			orQuit(err)

			ca := ndau.NewSetValidation(
				acct.Address,
				acct.Ownership.Public,
				[]signature.PublicKey{newKeys.Public},
				acct.ValidationScript,
				sequence(conf, acct.Address),
				acct.Ownership.Private,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), ca)

			// only persist this change if there was no error
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				acct.Validation = []config.Keypair{*newKeys}
				conf.SetAccount(*acct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "account set-validation")
		}
	}
}
