package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/tool"
)

func getCmdSeq(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf("%s", getNamespaceSpec())

		getNs := getNamespaceClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			ns := getNs(config)
			seq, response, err := tool.Sequence(tmnode(config.Node), ns)

			if err == nil {
				if *verbose {
					fmt.Printf("%x: %d\n", ns, seq)
				} else {
					fmt.Println(seq)
				}
			}
			finish(*verbose, response, err, "seq")
		}
	}
}
