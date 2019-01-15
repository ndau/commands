package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	cli "github.com/jawher/mow.cli"
	generator "github.com/oneiro-ndev/chaos_genesis/pkg/genesis.generator"
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
		bpcs     = cmd.StringOpt("bpc", "", "base64-encoded BPC public key")
	)

	cmd.Spec = "ASSC [--bpc]"

	cmd.Action = func() {
		if asscpath == nil || len(*asscpath) == 0 {
			orQuit(errors.New("path to associated data must be set"))
		}

		asscFile := make(generator.AssociatedFile)
		_, err := toml.DecodeFile(*asscpath, &asscFile)
		orQuit(err)

		if len(asscFile) == 1 && (bpcs == nil || len(*bpcs) == 0) {
			fmt.Println("bpcs not set; inferring from single value in assc datafile")
			// we can update bpcs with the only key in the asscfile
			for k := range asscFile {
				bpcs = &k
			}
		}

		if bpcs == nil || len(*bpcs) == 0 {
			orQuit(errors.New("if assc data file has other than 1 key, bpcs must be set"))
		}

		bpc, err := base64.StdEncoding.DecodeString(*bpcs)
		orQuit(errors.Wrap(err, "decoding bpc public key"))

		conf, err := config.Load(config.GetConfigPath())
		orQuit(errors.Wrap(err, "loading existing config"))

		err = conf.UpdateFrom(*asscpath, bpc)
		orQuit(errors.Wrap(err, "updating config"))

		err = conf.Save()
		orQuit(errors.Wrap(err, "failed to save configuration"))
	}
}
