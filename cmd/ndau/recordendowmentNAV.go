package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getRecordEndowmentNAV(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s",
			getNanocentSpec(),
		)

		getNanocent := getNanocentClosure(cmd)

		cmd.Action = func() {
			nanocentQty := getNanocent()

			if *verbose {
				fmt.Printf(
					"RecordEndowmentNAV %d nanocents\n",
					nanocentQty,
				)
			}

			conf := getConfig()
			if conf.RFE == nil {
				orQuit(errors.New("RecordEndowmentNAV keys not set in config"))
			}

			// construct the RecordEndowmentNAV
			RecordEndowmentNAV := ndau.NewRecordEndowmentNAV(
				nanocentQty,
				sequence(conf, conf.RFE.Address),
				conf.RFE.Keys...,
			)

			tresp, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), RecordEndowmentNAV)
			finish(*verbose, tresp, err, "record-price")
		}
	}
}
