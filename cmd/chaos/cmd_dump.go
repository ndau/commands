package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/tool"
)

func getCmdDump(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s",
			getNamespaceSpec(),
			getHeightSpec(),
			getDumpSpec(),
		)

		getHeight := getHeightClosure(cmd)
		getNamespace := getNamespaceClosure(cmd)
		dump := getDumpClosure(cmd)

		cmd.Action = func() {
			config := getConfig()

			value, result, err := tool.DumpNamespacedAt(
				tmnode(config.Node),
				getNamespace(config),
				getHeight(),
			)

			if err == nil {
				dump(os.Stdout, value)
			}
			finish(*verbose, result, err, "dump")
		}
	}
}
