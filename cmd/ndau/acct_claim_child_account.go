package main

import (
	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
	rpc "github.com/tendermint/tendermint/rpc/core/types"
)

func getAccountClaimChild(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME CHILD_NAME [-p=<CHILD_SETTLEMENT_PERIOD>] [--hd]"

		var (
			parentName = cmd.StringArg("NAME", "", "Name of parent account")
			childName  = cmd.StringArg("CHILD_NAME", "", "Name of child account to claim")
			period     = cmd.StringOpt("p period", "", "Initial settlement period for the child account (ndaumath types.ParseDuration format)")
			hd         = cmd.BoolOpt("hd", false, "Generate an HD key for the child account")
		)

		cmd.Action = func() {
			conf := getConfig()

			// Validate the child settlement period.
			var childSettlementPeriod math.Duration
			if period == nil || *period == "" {
				childSettlementPeriod = -math.Duration(1) // Use the default settlement period.
			} else {
				dur, err := math.ParseDuration(*period)
				orQuit(errors.Wrap(err, "Invalid child settlement period"))
				childSettlementPeriod = dur
			}

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

			cca := ndau.NewCreateChildAccount(
				parentAcct.Address,
				childAcct.Address,
				childAcct.Ownership.Public,
				childAcct.Ownership.Private.Sign([]byte(childAcct.Address.String())),
				childSettlementPeriod,
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
