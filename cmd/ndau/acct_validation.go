package main

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

func getAccountValidation(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME"

		var name = cmd.StringArg("NAME", "", "Name of account to change")

		cmd.Command(
			"reset",
			"generate a new transfer key which replaces all current transfer keys",
			getReset(verbose, name, keys),
		)

		cmd.Command(
			"add",
			"add a new transfer key to this account",
			getAdd(verbose, name, keys),
		)

		cmd.Command(
			"recover",
			"add a recovered key to this account from its path, not touching the blockchain",
			getRecover(verbose, name),
		)

		cmd.Command(
			"set-script",
			"set validation script for this account",
			getSetScript(verbose, name, keys),
		)
	}
}

func getReset(verbose *bool, name *string, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = getKeypathSpec(true)

		getKeypath := getKeypathClosure(cmd, true)

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			if len(acct.Transfer) == 0 {
				orQuit(errors.New("account is not yet claimed"))
			}

			keypath := getKeypath()
			newkeys, err := acct.MakeTransferKey(&keypath)
			orQuit(errors.Wrap(err, "failed to generate new transfer key"))

			cv := ndau.NewChangeValidation(
				acct.Address,
				[]signature.PublicKey{newkeys.Public},
				acct.ValidationScript,
				sequence(conf, acct.Address),
				acct.TransferPrivateK(keys),
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), &cv)

			// only persist this change if there was no error
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				acct.Transfer = []config.Keypair{*newkeys}
				conf.SetAccount(*acct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "account validation reset")
		}
	}
}

func getAdd(verbose *bool, name *string, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = getKeypathSpec(true)

		getKeypath := getKeypathClosure(cmd, true)

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			if len(acct.Transfer) == 0 {
				orQuit(errors.New("account is not yet claimed"))
			}

			keypath := getKeypath()
			newkeys, err := acct.MakeTransferKey(&keypath)
			orQuit(errors.Wrap(err, "failed to generate new transfer key"))

			cv := ndau.NewChangeValidation(
				acct.Address,
				append(acct.TransferPublic(), newkeys.Public),
				acct.ValidationScript,
				sequence(conf, acct.Address),
				acct.TransferPrivateK(keys),
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), &cv)

			// only persist this change if there was no error
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				acct.Transfer = append(acct.Transfer, *newkeys)
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
			newkeys, err := acct.MakeTransferKey(&keypath)
			orQuit(errors.Wrap(err, "failed to key from path"))

			acct.Transfer = []config.Keypair{*newkeys}
			conf.SetAccount(*acct)
			err = conf.Save()
			orQuit(errors.Wrap(err, "saving config"))

			finish(*verbose, nil, err, "account validation recover")
		}
	}
}

func getSetScript(verbose *bool, name *string, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[SCRIPT]"

		scriptB64 := cmd.StringArg("SCRIPT", "", "base64-encoded validation script")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			if len(acct.Transfer) == 0 {
				orQuit(errors.New("account is not yet claimed"))
			}

			script, err := base64.RawStdEncoding.DecodeString(*scriptB64)
			orQuit(err)

			if *verbose {
				fmt.Printf("Script b64: %s\n       hex: %x\n", *scriptB64, script)
			}

			cv := ndau.NewChangeValidation(
				acct.Address,
				acct.TransferPublic(),
				script,
				sequence(conf, acct.Address),
				acct.TransferPrivateK(keys),
			)

			if *verbose {
				fmt.Printf("%#v\n", cv)
			}

			resp, err := tool.SendCommit(tmnode(conf.Node), &cv)

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
