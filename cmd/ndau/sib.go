package main

import (
	"encoding/json"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getSIB(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			info, resp, err := tool.GetSIB(tmnode(config.Node, nil, nil))
			if err == nil {
				var jsi []byte
				jsi, err = json.MarshalIndent(info, "", "  ")
				fmt.Println(string(jsi))
			}

			finish(*verbose, resp, err, "sib")
		}
	}
}
