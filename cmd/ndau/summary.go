package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
)

func getSummary(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		hgt := cmd.BoolOpt("h height", false, "when set, emit only the block height")
		acct := cmd.BoolOpt("a accounts", false, "when set, emit only the number of accounts")
		tot := cmd.BoolOpt("t total", false, "when set, emit only the total number of ndau")
		napu := cmd.BoolOpt("n napu", false, "when set, emit only the total number of napu")

		cmd.Spec = "[ -h | -a | -t | -n ]"

		cmd.Action = func() {
			config := getConfig()
			info, _, err := tool.GetSummary(tmnode(config.Node))

			// if none of them are set, turn on verbose if it's not on and set the first 3
			if !*hgt && !*acct && !*tot && !*napu {
				*verbose = true
				*hgt = true
				*acct = true
				*tot = true
			}

			if *hgt && !*verbose {
				f := "%d\n"
				fmt.Printf(f, info.BlockHeight)
			}

			if *acct && !*verbose {
				f := "%d\n"
				fmt.Printf(f, info.NumAccounts)
			}

			if *tot && !*verbose {
				f := "%f\n"
				fmt.Printf(f, float64(info.TotalNdau)/constants.NapuPerNdau)
			}

			if *napu && !*verbose {
				f := "%d\n"
				fmt.Printf(f, info.TotalNdau)
			}

			finish(*verbose, info, err, "summary")
		}
	}
}
