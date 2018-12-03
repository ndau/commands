package main

import (
	"os"

	cli "github.com/jawher/mow.cli"
	util "github.com/oneiro-ndev/genesis/pkg/cli.util"
	"github.com/oneiro-ndev/genesis/pkg/config"
	genesis "github.com/oneiro-ndev/genesis/pkg/genesis"
)

func main() {
	app := cli.App("update_genesis_json", "update a chain's genesis.json file")

	name := app.StringArg("NAME", "", "Name of chain ('chaos', 'order', 'ndau', etc.)")
	path := app.StringArg("PATH", "", "Path to genesis.json")

	app.Action = func() {
		configPath := config.DefaultConfigPath(util.GetNdauhome())
		err := config.WithConfig(configPath, func(conf *config.Config) error {
			return genesis.ProcessGenesisJSON(conf, *name, *path)
		})
		util.Check(err)
	}

	app.Run(os.Args)
}
