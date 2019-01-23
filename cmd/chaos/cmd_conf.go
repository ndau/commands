package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/tool"
	"github.com/pkg/errors"
)

func getCmdConf(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[ADDR] [--ndau]"

		var (
			addr = cmd.StringArg("ADDR", tool.DefaultAddress, "rpc address of chaos node")
			ndau = cmd.StringOpt("N ndau", "", "rpc address of ndau node")
		)

		cmd.Action = func() {
			config, err := tool.Load()
			if err != nil && os.IsNotExist(err) {
				config = tool.NewConfig(*addr)
			} else {
				config.Node = *addr
			}
			if ndau != nil && *ndau != "" {
				config.NdauAddress = ndau
			}
			err = config.Save()
			orQuit(errors.Wrap(err, "Failed to save configuration"))
		}
	}
}

func getCmdConfPath() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			fmt.Println(tool.GetConfigPath())
		}
	}
}
