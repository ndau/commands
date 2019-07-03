package main

import (
	"encoding/base64"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getSetStakeRules(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		var (
			name      = cmd.StringArg("NAME", "", "Name of account to lock")
			scriptB64 = cmd.StringArg("RULES", "", "base64-encoded stake rules script")
		)

		cmd.Spec = "NAME RULES"

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			rules, err := base64.StdEncoding.DecodeString(*scriptB64)
			orQuit(errors.Wrap(err, "decoding chaincode"))

			if *verbose {
				fmt.Printf("Script b64: %s\n       hex: %x\n", *scriptB64, rules)
			}

			tx := ndau.NewSetStakeRules(
				acct.Address,
				rules,
				sequence(conf, acct.Address),
				acct.ValidationPrivateK(*keys)...,
			)

			if *verbose {
				fmt.Printf("%#v\n", tx)
			}

			resp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), tx)

			finish(*verbose, resp, err, "account set-stake-rules")
		}
	}
}
