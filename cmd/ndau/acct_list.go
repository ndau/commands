package main

import (
	"os"

	cli "github.com/jawher/mow.cli"
)

func getAccountList(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			config.EmitAccounts(os.Stdout)
		}
	}
}
