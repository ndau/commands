package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/pkg/errors"
)

func getRfe(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s",
			getNdauSpec(),
			getAddressSpec(""),
		)

		getNdau := getNdauClosure(cmd)
		getAddress := getAddressClosure(cmd, "")

		cmd.Action = func() {
			ndauQty := getNdau()
			address := getAddress()

			if *verbose {
				fmt.Printf("Release from endowment: %s ndau to %s\n", ndauQty, address)
			}

			conf := getConfig()
			if conf.RFE == nil {
				orQuit(errors.New("RFE data not set in tool config"))
			}

			keys := config.FilterK(conf.RFE.Keys, keys)

			rfe := ndau.NewReleaseFromEndowment(
				ndauQty,
				address,
				sequence(conf, conf.RFE.Address),
				keys,
			)

			result, err := tool.SendCommit(tmnode(conf.Node), &rfe)
			finish(*verbose, result, err, "rfe")
		}
	}
}
