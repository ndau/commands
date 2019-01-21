package main

import (
	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/tool"
)

func getCmdInfo() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			info, err := tool.Info(tmnode(config.Node))
			// the whole point of this command is to get this information;
			// it makes no sense to require the verbose flag in this case
			finish(true, info, err, "info")
		}
	}
}
