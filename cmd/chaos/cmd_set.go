package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	twrite "github.com/oneiro-ndev/chaos/pkg/tool.write"
)

func getCmdSet(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"NAME %s %s",
			getKeySpec(),
			getValueSpec(),
		)

		var (
			name     = cmd.StringArg("NAME", "", "Name of identity to use")
			getKey   = getKeyClosure(cmd, verbose)
			getValue = getValueClosure(cmd, verbose)
		)

		cmd.Action = func() {
			config := getConfig()
			result, err := twrite.Set(tmnode(config.Node), *name, config, getKey(), getValue())
			finish(*verbose, result, err, "set")
		}
	}
}
