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
	"encoding/base64"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
	rpc "github.com/tendermint/tendermint/rpc/core/types"
)

func getAccountValidation(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME"

		var name = cmd.StringArg("NAME", "", "Name of account to change")

		cmd.Command(
			"reset",
			"generate a new validation key which replaces all current validation keys",
			getReset(verbose, name, keys, emitJSON, compact),
		)

		cmd.Command(
			"add",
			"add a new validation key to this account",
			getAdd(verbose, name, keys, emitJSON, compact),
		)

		cmd.Command(
			"recover",
			"add a recovered key to this account from its path, not touching the blockchain",
			getRecover(verbose, name),
		)

		cmd.Command(
			"set-script",
			"set validation script for this account",
			getSetScript(verbose, name, keys, emitJSON, compact),
		)
	}
}

func getReset(verbose *bool, name *string, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = getKeypathSpec(true)

		getKeypath := getKeypathClosure(cmd, true)

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			if len(acct.Validation) == 0 {
				orQuit(errors.New("account doesn't have validation keys"))
			}

			keypath := getKeypath()
			newkeys, err := acct.MakeValidationKey(&keypath)
			orQuit(errors.Wrap(err, "failed to generate new validation key"))

			cv := ndau.NewChangeValidation(
				acct.Address,
				[]signature.PublicKey{newkeys.Public},
				acct.ValidationScript,
				sequence(conf, acct.Address),
				acct.ValidationPrivateK(*keys)...,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), cv)

			// only persist this change if there was no error
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				acct.Validation = []config.Keypair{*newkeys}
				conf.SetAccount(*acct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "account validation reset")
		}
	}
}

func getAdd(verbose *bool, name *string, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = getKeypathSpec(true)

		getKeypath := getKeypathClosure(cmd, true)

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			if len(acct.Validation) == 0 {
				orQuit(errors.New("account doesn't have validation keys"))
			}

			keypath := getKeypath()
			newkeys, err := acct.MakeValidationKey(&keypath)
			orQuit(errors.Wrap(err, "failed to generate new validation key"))

			cv := ndau.NewChangeValidation(
				acct.Address,
				append(acct.ValidationPublic(), newkeys.Public),
				acct.ValidationScript,
				sequence(conf, acct.Address),
				acct.ValidationPrivateK(*keys)...,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), cv)

			// only persist this change if there was no error
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				acct.Validation = append(acct.Validation, *newkeys)
				conf.SetAccount(*acct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "account validation add")
		}
	}
}

func getRecover(verbose *bool, name *string) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = getKeypathSpec(false)

		getKeypath := getKeypathClosure(cmd, false)

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			keypath := getKeypath()
			newkeys, err := acct.MakeValidationKey(&keypath)
			orQuit(errors.Wrap(err, "failed to key from path"))

			acct.Validation = []config.Keypair{*newkeys}
			conf.SetAccount(*acct)
			err = conf.Save()
			orQuit(errors.Wrap(err, "saving config"))

			finish(*verbose, nil, err, "account validation recover")
		}
	}
}

func getSetScript(verbose *bool, name *string, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[SCRIPT]"

		scriptB64 := cmd.StringArg("SCRIPT", "", "base64-encoded validation script")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			if len(acct.Validation) == 0 {
				orQuit(errors.New("account doesn't have validation keys"))
			}

			script, err := base64.RawStdEncoding.DecodeString(*scriptB64)
			orQuit(err)

			if *verbose {
				fmt.Printf("Script b64: %s\n       hex: %x\n", *scriptB64, script)
			}

			cv := ndau.NewChangeValidation(
				acct.Address,
				acct.ValidationPublic(),
				script,
				sequence(conf, acct.Address),
				acct.ValidationPrivateK(*keys)...,
			)

			if *verbose {
				fmt.Printf("%#v\n", cv)
			}

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), cv)

			// only persist this change if there was no error
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				acct.ValidationScript = script
				conf.SetAccount(*acct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "account validation add")
		}
	}
}
