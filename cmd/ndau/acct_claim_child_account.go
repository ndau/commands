package main

import (
	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
	rpc "github.com/tendermint/tendermint/rpc/core/types"
)

func getAccountClaimChild(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME CHILD_NAME [--hd]"

		var (
			parentName = cmd.StringArg("NAME", "", "Name of parent account")
			childName  = cmd.StringArg("CHILD_NAME", "", "Name of child account to claim")
			hd         = cmd.BoolOpt("hd", false, "when set, generate a HD key for this child account")
		)

		cmd.Action = func() {
			conf := getConfig()

			// Ensure the parent account exists in the config already.
			parentAcct, hasAcct := conf.Accounts[*parentName]
			if !hasAcct {
				orQuit(errors.New("Parent account does not exist"))
			}

			// Transaction validation would catch this, but it's helpful to catch it early.
			if len(parentAcct.Transfer) == 0 {
				orQuit(errors.New("Parent account is not yet claimed"))
			}

			// Ensure the child account does not exist in the config yet.
			_, hasAcct = conf.Accounts[*childName]
			if hasAcct {
				orQuit(errors.New("Child account already exists"))
			}

			// Add the new child account to the config.
			err := conf.CreateAccount(*childName, *hd)
			orQuit(errors.Wrap(err, "Failed to create child identity"))

			// Get the non-nil child account now that it exists in the config.
			childAcct, hasAcct := conf.Accounts[*childName]
			if !hasAcct {
				orQuit(errors.New("Child account does not exists"))
			}

			// Transaction validation would catch this, but it's helpful to catch it early.
			if len(childAcct.Transfer) != 0 {
				orQuit(errors.New("Child account is already claimed"))
			}

			newChildKeys, err := childAcct.MakeTransferKey(nil)
			orQuit(err)

			cca := ndau.NewClaimChildAccount(
				parentAcct.Address,
				childAcct.Address,
				childAcct.Ownership.Public,
				childAcct.Ownership.Private.Sign([]byte(childAcct.Address.String())),
				[]signature.PublicKey{newChildKeys.Public},
				childAcct.ValidationScript,
				sequence(conf, parentAcct.Address),
				parentAcct.TransferPrivateK(*keys)...,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), cca)

			// Only persist this change if there was no error.
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				childAcct.Transfer = []config.Keypair{*newChildKeys}
				conf.SetAccount(*childAcct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "account claim child")
		}
	}
}
