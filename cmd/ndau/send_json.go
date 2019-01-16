package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getSendJSON(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s",
			getJSONTXSpec(),
		)

		getJSONTX := getJSONTXClosure(cmd)

		cmd.Action = func() {
			tx := getJSONTX()
			conf := getConfig()
			resp, err := tool.SendCommit(tmnode(conf.Node, nil, nil), tx)
			finish(*verbose, resp, err, "send-json")
		}
	}
}
