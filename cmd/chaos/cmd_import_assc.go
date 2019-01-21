package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

func getCmdImportAssc() func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[IDENTITY] ASSC [--bpc]"

		var (
			idName   = cmd.StringArg("IDENTITY", "sysvar", "name of identity to set from bpc data in assc.toml")
			asscPath = cmd.StringArg("ASSC", "", "path to assc.toml")
			bpc      = cmd.StringOpt("bpc", "", "base64-encoded bytes of bpc public key")
		)

		cmd.Action = func() {
			bpcsmap := make(map[string]interface{})
			_, err := toml.DecodeFile(*asscPath, &bpcsmap)
			orQuit(err)

			if bpc == nil || len(*bpc) == 0 {
				if len(bpcsmap) == 1 {
					// we want the only entry
					for k := range bpcsmap {
						bpc = &k
					}
				} else {
					orQuit(errors.New("assc.toml ambiguous: must manually specify bpc public key"))
				}
			}

			keysmap := bpcsmap[*bpc].(map[string]interface{})

			publicS, ok := keysmap["BPCPublic"]
			if !ok {
				orQuit(errors.New("BPCPublic key not found"))
			}
			public, err := signature.ParsePublicKey(publicS.(string))
			orQuit(errors.Wrap(err, "BPCPublic"))

			privateS, ok := keysmap["BPCPrivate"]
			if !ok {
				orQuit(errors.New("BPCPrivate key not found"))
			}
			private, err := signature.ParsePrivateKey(privateS.(string))
			orQuit(errors.Wrap(err, "BPCPrivate"))

			config, err := tool.Load()
			orQuit(err)

			config.Identities[*idName] = tool.Identity{
				Name: *idName,
				Chaos: tool.Keypair{
					Public:  *public,
					Private: private,
				},
			}
			orQuit(config.Save())

			// temporary code
			// see https://github.com/oneiro-ndev/commands/pull/64/files#r248228258
			fmt.Println("IMPORTANT: you are not yet done")
			fmt.Println()
			fmt.Println("You must edit the configuration document:")
			fmt.Println("  i.e.:  $ nano $(./chaos conf-path)")
			fmt.Println()
			fmt.Println("- find the ", *idName, " identity")
			fmt.Println("- add to that identity an [identities.ndau] table")
			fmt.Println("- add to that table: address: \"ndbmgby86qw9bds9f8wrzut5zrbxuehum5kvgz9sns9hgknh\"")
			fmt.Println("- add to that table: [[identities.ndau.keys]]")
			fmt.Println("- get the public and private keys for BPC operations from 1password")
			fmt.Println("- add those keys to this keys section")
			fmt.Println()
			fmt.Println("Once this is complete, you should be able to update system variables")
		}
	}
}
