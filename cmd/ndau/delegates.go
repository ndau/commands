package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getDelegates(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			conf := getConfig()
			delegates, resp, err := tool.GetDelegates(tmnode(conf.Node, nil, nil))

			// errors get logged in the Info field
			fmt.Fprint(os.Stderr, resp.Response.GetInfo())

			// display the delegates
			for node, delegated := range delegates {
				fmt.Printf("%s:\n", node)
				for _, d := range delegated {
					fmt.Printf("  %s\n", d)
				}
			}

			finish(*verbose, resp, err, "show-delegates")
		}
	}
}
