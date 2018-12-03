package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getTransfer(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s",
			getNdauSpec(),
			getAddressSpec("FROM"),
			getAddressSpec("TO"),
		)

		getNdau := getNdauClosure(cmd)
		getFrom := getAddressClosure(cmd, "FROM")
		getTo := getAddressClosure(cmd, "TO")

		cmd.Action = func() {
			ndauQty := getNdau()
			from := getFrom()
			to := getTo()

			if *verbose {
				fmt.Printf(
					"Transfer %s ndau from %s to %s\n",
					ndauQty, from, to,
				)
			}

			conf := getConfig()

			// ensure we know the private transfer key of this account
			fromAcct, hasAcct := conf.Accounts[from.String()]
			if !hasAcct {
				orQuit(fmt.Errorf("From account '%s' not found", fromAcct.Name))
			}
			if fromAcct.Transfer == nil {
				orQuit(fmt.Errorf("From acct transfer key not set"))
			}

			// construct the transfer
			transfer, err := ndau.NewTransfer(
				from, to,
				ndauQty,
				sequence(conf, from),
				fromAcct.TransferPrivateK(keys),
			)
			orQuit(errors.Wrap(err, "Failed to construct transfer"))

			tresp, err := tool.SendCommit(tmnode(conf.Node), transfer)
			finish(*verbose, tresp, err, "transfer")
		}
	}
}
