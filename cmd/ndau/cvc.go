package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/pkg/errors"
)

func getCVC(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME POWER"

		var (
			name  = cmd.StringArg("NAME", "", "Name of node to register")
			power = cmd.IntArg("POWER", 0, "power to assign to this node")
		)

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}

			if *power < 0 {
				orQuit(errors.New("cvc POWER must be > 0"))
			}

			if *verbose {
				fmt.Printf("CommandValidatorChange: %s Power %d\n", acct.Address, *power)
			}

			if conf.CVC == nil {
				orQuit(errors.New("CVC data not set in tool config"))
			}

			fkeys := config.FilterK(conf.CVC.Keys, *keys)

			cvc := ndau.NewCommandValidatorChange(
				acct.Address, int64(*power),
				sequence(conf, conf.CVC.Address),
				fkeys...,
			)

			result, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), cvc)
			finish(*verbose, result, err, "cvc")
		}
	}
}
