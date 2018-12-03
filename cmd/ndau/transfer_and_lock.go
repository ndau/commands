package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getTransferAndLock(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s %s",
			getNdauSpec(),
			getAddressSpec("FROM"),
			getAddressSpec("TO"),
			getDurationSpec(),
		)

		getNdau := getNdauClosure(cmd)
		getFrom := getAddressClosure(cmd, "FROM")
		getTo := getAddressClosure(cmd, "TO")
		getDuration := getDurationClosure(cmd)

		cmd.Action = func() {
			ndauQty := getNdau()
			from := getFrom()
			to := getTo()
			duration := getDuration()

			if *verbose {
				fmt.Printf(
					"Transfer %s ndau from %s to %s, locking for %s\n",
					ndauQty, from, to, duration,
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
			transfer, err := ndau.NewTransferAndLock(
				from, to,
				ndauQty,
				duration,
				sequence(conf, from),
				fromAcct.TransferPrivateK(keys),
			)
			orQuit(errors.Wrap(err, "Failed to construct transferand lock"))

			tresp, err := tool.SendCommit(tmnode(conf.Node), transfer)
			finish(*verbose, tresp, err, "transferandlock")
		}
	}
}
