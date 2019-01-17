package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
)

func getAccountDestroy(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME [-F]"

		var (
			name  = cmd.StringArg("NAME", "", "name of account to be destroyed")
			force = cmd.BoolOpt("F force", false, "proceed without manual confirmation")
		)

		cmd.Action = func() {
			conf := getConfig()

			acct, ok := conf.Accounts[*name]
			if !ok {
				orQuit(fmt.Errorf("unknown acct: \"%s\"", *name))
			}

			if !*force {
				fmt.Println("WARNING: this will permanently remove all local knowledge of this account")
				fmt.Println()
				fmt.Println("If the account is based on a known 12-word phrase, it can be recovered.")
				fmt.Println("Otherwise, there is no way to regain access to this account.")
				fmt.Println()
				fmt.Print("To proceed, re-enter this account's name: ")

				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				input := scanner.Text()

				if strings.TrimSpace(input) != *name {
					fmt.Println("mismatch; aborting")
					os.Exit(1)
				}
			}

			delete(conf.Accounts, acct.Name)
			delete(conf.Accounts, acct.Address.String())

			orQuit(conf.Save())

			if *verbose {
				fmt.Println("destroyed acct:", *name)
			}
		}
	}
}
