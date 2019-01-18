package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
)

func getSignableBytes(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s",
			getEmitBytesSpec(),
			getJSONTXSpec(),
		)

		emitBytes := getEmitBytesClosure(cmd)
		getJSONTX := getJSONTXClosure(cmd)

		cmd.Action = func() {
			tx := getJSONTX()
			emitBytes(tx.SignableBytes())
		}
	}
}
