package main

import (
	"encoding/base64"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/chaos/ns"
	"github.com/oneiro-ndev/chaos/pkg/tool"
)

func getCmdGetNS(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s",
			getHeightSpec(),
		)

		getHeight := getHeightClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			rim := config.ReverseIdentityMap()
			height := getHeight()

			namespaces, result, err := tool.GetNamespacesAt(tmnode(config.Node), height)

			if err == nil {
				for _, namespace := range namespaces {
					if !ns.IsSpecial(namespace) {
						b64 := base64.StdEncoding.EncodeToString(namespace)
						if name, hasName := rim[string(namespace)]; hasName {
							fmt.Printf("%s (%s)\n", b64, name)
						} else {
							fmt.Println(b64)
						}
					}
				}
			}

			finish(*verbose, result, err, "get-ns")
		}
	}
}
