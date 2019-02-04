package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/pkg/errors"
)

func getIssue(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = getNdauSpec()

		getNdau := getNdauClosure(cmd)

		cmd.Action = func() {
			ndauQty := getNdau()

			if *verbose {
				fmt.Printf("Issue: %s ndau\n", ndauQty)
			}

			conf := getConfig()
			if conf.RFE == nil {
				orQuit(errors.New("RFE data (for issuance) not set in tool config"))
			}

			keys := config.FilterK(conf.RFE.Keys, *keys)

			issue := ndau.NewIssue(
				ndauQty,
				sequence(conf, conf.RFE.Address),
				keys...,
			)

			result, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), issue)
			finish(*verbose, result, err, "issue")
		}
	}
}
