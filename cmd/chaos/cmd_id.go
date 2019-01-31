package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/tool"
	ntconf "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/pkg/errors"
)

func getCmdID() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command("list", "list known identities", getCmdIDList())
		cmd.Command("new", "create a new identity", getCmdIDName())
		cmd.Command("copy-keys-from", "copy ndau keys to local config", getCmdIDCopyKeysFrom())
	}
}

func getCmdIDList() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			config.EmitIdentities(os.Stdout)
		}
	}
}

func getCmdIDName() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME"

		var (
			name = cmd.StringArg("NAME", "", "Name to associate with the new identity")
		)

		cmd.Action = func() {
			config := getConfig()
			err := config.CreateIdentity(*name, os.Stdout)
			orQuit(errors.Wrap(err, "Failed to create identity"))
			orQuit(config.Save())
		}
	}
}

func getCmdIDCopyKeysFrom() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME [NDAU_NAME] [-p=<path/to/ndautool.toml>]"

		var (
			name     = cmd.StringArg("NAME", "", "chaos identity name")
			ndauName = cmd.StringArg("NDAU_NAME", "", "ndau identity name (default: same as NAME)")
			ntPath   = cmd.StringOpt("p ndautool", ntconf.GetConfigPath(), "path to ndautool.toml")
		)

		cmd.Action = func() {
			conf := getConfig()
			id, ok := conf.Identities[*name]
			if !ok {
				orQuit(fmt.Errorf("%s is not a known identity", *name))
			}

			if len(*ndauName) == 0 {
				ndauName = name
			}

			ndauConfig, err := ntconf.Load(*ntPath)
			orQuit(err)

			nid, ok := ndauConfig.Accounts[*ndauName]
			if !ok {
				orQuit(fmt.Errorf("%s is not a known ndau identity", *ndauName))
			}

			if len(nid.Transfer) == 0 {
				orQuit(fmt.Errorf("%s has no transfer keys", *ndauName))
			}

			id.Ndau = &tool.NdauAccount{
				Address: nid.Address,
			}

			for _, trKeys := range nid.Transfer {
				id.Ndau.Keys = append(
					id.Ndau.Keys,
					tool.Keypair{
						Public:  trKeys.Public,
						Private: &trKeys.Private,
					},
				)
			}

			conf.Identities[*name] = id

			orQuit(conf.Save())
		}
	}
}
