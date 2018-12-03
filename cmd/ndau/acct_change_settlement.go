package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getAccountChangeSettlement(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"NAME %s",
			getDurationSpec(),
		)

		var name = cmd.StringArg("NAME", "", "Name of account of which to change settlement period")
		getDuration := getDurationClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			duration := getDuration()

			ad, hasAd := config.Accounts[*name]
			if !hasAd {
				orQuit(errors.New("No such account found"))
			}
			if ad.Transfer == nil {
				orQuit(errors.New("Address transfer key not set"))
			}

			cep, err := ndau.NewChangeSettlementPeriod(
				ad.Address,
				duration,
				sequence(config, ad.Address),
				ad.TransferPrivateK(keys),
			)
			orQuit(errors.Wrap(err, "Creating ChangeEscrowPeriod transaction"))

			if *verbose {
				fmt.Printf(
					"Change Escrow Period for %s (%s) to %s\n",
					*name,
					ad.Address,
					duration,
				)
			}

			resp, err := tool.SendCommit(tmnode(config.Node), &cep)
			finish(*verbose, resp, err, "change-settlement-period")
		}
	}
}
