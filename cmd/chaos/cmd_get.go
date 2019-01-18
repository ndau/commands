package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/tool"
)

func getCmdGet(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s %s",
			getNamespaceSpec(),
			getHeightSpec(),
			getKeySpec(),
			getEmitSpec(),
		)

		getNs := getNamespaceClosure(cmd)
		getHeight := getHeightClosure(cmd)
		getKey := getKeyClosure(cmd)
		emit := getEmitClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			namespace := getNs(config)

			value, result, err := tool.GetNamespacedAt(
				tmnode(config.Node), namespace,
				getKey(), getHeight(),
			)
			if err == nil {
				emit(value)
			}

			finish(*verbose, result, err, "get")
		}
	}
}
