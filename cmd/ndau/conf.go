package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"

	cli "github.com/jawher/mow.cli"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/pkg/errors"
)

func getConf(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[ADDR]"

		var addr = cmd.StringArg("ADDR", config.DefaultAddress, "Address of node to connect to")

		cmd.Action = func() {
			conf, err := config.Load(config.GetConfigPath())
			if err != nil && os.IsNotExist(err) {
				conf = config.NewConfig(*addr)
			} else {
				orQuit(errors.Wrap(err, "loading config"))
				conf.Node = *addr
			}
			err = conf.Save()
			orQuit(errors.Wrap(err, "Failed to save configuration"))
		}

		cmd.Command("update-from", "update the config from an associated-data file", confUpdateFrom)
	}
}

func confPath(cmd *cli.Cmd) {
	cmd.Action = func() {
		fmt.Println(config.GetConfigPath())
	}
}

func confUpdateFrom(cmd *cli.Cmd) {
	var (
		asscpath = cmd.StringArg("ASSC", "", "Path to associated data file")
	)

	cmd.Spec = "ASSC"

	cmd.Action = func() {
		if asscpath == nil || len(*asscpath) == 0 {
			orQuit(errors.New("path to associated data must be set"))
		}

		conf, err := config.Load(config.GetConfigPath())
		orQuit(errors.Wrap(err, "loading existing config"))

		err = conf.UpdateFrom(*asscpath)
		orQuit(errors.Wrap(err, "updating config"))

		// create a normal account for the node rules account
		var apd map[string]interface{}
		_, err = toml.DecodeFile(*asscpath, &apd)
		orQuit(errors.Wrap(err, "decoding associated data"))
		sa, err := config.SysAccountFromAssc(apd, sv.NodeRulesAccount)
		orQuit(errors.Wrap(err, "getting node rules associated data"))
		conf.SetAccount(config.Account{
			Name:    "node-rules",
			Address: sa.Address,
			Validation: []config.Keypair{
				config.Keypair{
					Private: sa.Keys[0],
				},
			},
		})

		err = conf.Save()
		orQuit(errors.Wrap(err, "failed to save configuration"))
	}
}
