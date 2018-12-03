package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getInfo(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		key := cmd.BoolOpt("k key", false, "when set, emit the public key of the connected node")
		emithex := cmd.BoolOpt("x hex", false, "when set, emit the key as hex instead of base64")
		pwr := cmd.BoolOpt("p power", false, "when set, emit the power of the connected node")

		cmd.Spec = "[-k [-x]] [-p]"

		cmd.Action = func() {
			config := getConfig()
			info, err := tool.Info(tmnode(config.Node))

			if *key {
				b := info.ValidatorInfo.PubKey.Bytes()
				var p string
				if *emithex {
					p = hex.EncodeToString(b)
				} else {
					p = base64.RawStdEncoding.EncodeToString(b)
				}
				fmt.Println(p)
			}
			if *pwr {
				fmt.Println(info.ValidatorInfo.VotingPower)
			}
			if !(*key || *pwr) {
				*verbose = true
			}
			finish(*verbose, info, err, "info")
		}
	}
}
