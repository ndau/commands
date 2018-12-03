package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	v "github.com/oneiro-ndev/ndau/pkg/version"
	"github.com/pkg/errors"
)

func getVersion(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = v.Emit

		cmd.Command("remote", "emit the version of the connected node", getRemote(verbose))
		cmd.Command("check", "check that local and remote versions match", getCheck(verbose))
	}
}

func getRemote(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			version, resp, err := tool.Version(tmnode(config.Node))
			if version != "" {
				fmt.Println(version)
			}
			finish(*verbose, resp, err, "version remote")
		}
	}
}

func getCheck(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			local, err := v.Get()
			orQuit(err)

			config := getConfig()

			remote, resp, err := tool.Version(tmnode(config.Node))
			if err != nil {
				err = errors.Wrap(err, "fetching remote version")
				finish(*verbose, resp, err, "version check")
			}

			if local != remote {
				err = fmt.Errorf(
					"version mismatch: local %s; remote %s",
					local,
					remote,
				)
			}
			if *verbose && err == nil {
				fmt.Println("OK")
			}
			finish(*verbose, resp, err, "version check")
		}
	}
}
