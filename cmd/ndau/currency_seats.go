package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getCurrencySeats(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[N]"

		var (
			n = cmd.IntArg("N", 0, "return only the oldest N currency seats")
		)

		cmd.Action = func() {
			config := getConfig()
			seats, err := tool.GetCurrencySeats(tmnode(config.Node, nil, nil))
			orQuit(err)
			if n != nil && *n > 0 {
				if *n > len(seats) {
					*n = len(seats)
				}
				seats = seats[:*n]
			}

			for _, seat := range seats {
				fmt.Println(seat)
			}

			finish(*verbose, nil, err, "summary")
		}
	}
}
