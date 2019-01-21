package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/tool"
)

func getCmdHistory(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s",
			getNamespaceSpec(),
			getKeySpec(),
			getEmitHistorySpec(),
		)

		getNs := getNamespaceClosure(cmd)
		getKey := getKeyClosure(cmd, verbose)
		emit := getEmitHistoryClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			namespace := getNs(config)

			values, result, err := tool.HistoryNamespaced(
				tmnode(config.Node),
				namespace,
				getKey(),
				0,
				0,
			)
			if err == nil {
				emit(os.Stdout, values)
			}

			finish(*verbose, result, err, "history")
		}
	}
}
