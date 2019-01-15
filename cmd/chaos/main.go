package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/chaos/ns"
	"github.com/oneiro-ndev/chaos/pkg/tool"
	twrite "github.com/oneiro-ndev/chaos/pkg/tool.write"
	ntconf "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/pkg/errors"
)

func main() {
	app := cli.App("chaos", "interact with the chaos chain")

	app.Spec = "[-v]"

	var (
		verbose = app.BoolOpt("v verbose", false, "Emit detailed results from the chaos chain if set")
	)

	app.Command("conf", "perform initial configuration", func(cmd *cli.Cmd) {
		cmd.Spec = "[ADDR]"

		var addr = cmd.StringArg("ADDR", tool.DefaultAddress, "Address of node to connect to")

		cmd.Action = func() {
			config, err := tool.Load()
			if err != nil && os.IsNotExist(err) {
				config = tool.NewConfig(*addr)
			} else {
				config.Node = *addr
			}
			err = config.Save()
			orQuit(errors.Wrap(err, "Failed to save configuration"))
		}
	})

	app.Command("conf-path", "show location of config file", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			fmt.Println(tool.GetConfigPath())
		}
	})

	app.Command("import-assc", "import an assc.toml file", func(cmd *cli.Cmd) {
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
		}
	})

	app.Command("id", "manage identities", func(cmd *cli.Cmd) {
		cmd.Command("list", "list known identities", func(subcmd *cli.Cmd) {
			subcmd.Action = func() {
				config := getConfig()
				config.EmitIdentities(os.Stdout)
			}
		})

		cmd.Command("new", "create a new identity", func(subcmd *cli.Cmd) {
			subcmd.Spec = "NAME"

			var (
				name = subcmd.StringArg("NAME", "", "Name to associate with the new identity")
			)

			subcmd.Action = func() {
				config := getConfig()
				err := config.CreateIdentity(*name, os.Stdout)
				orQuit(errors.Wrap(err, "Failed to create identity"))
				orQuit(config.Save())
			}
		})

		cmd.Command("copy-keys-from", "copy ndau keys to local config", func(subcmd *cli.Cmd) {
			subcmd.Spec = "NAME [NDAU_NAME] [-p=<path/to/ndautool.toml>]"

			var (
				name     = subcmd.StringArg("NAME", "", "chaos identity name")
				ndauName = subcmd.StringArg("NDAU_NAME", "", "ndau identity name (default: same as NAME)")
				ntPath   = subcmd.StringOpt("p ndautool", ntconf.GetConfigPath(), "path to ndautool.toml")
			)

			subcmd.Action = func() {
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

				err = conf.Save()
			}
		})
	})

	app.Command("set", "set K-V pairs", func(cmd *cli.Cmd) {
		allowBinaryKey := true

		cmd.Spec = fmt.Sprintf(
			"NAME %s %s",
			getKeySpec(allowBinaryKey),
			getValueSpec(),
		)

		var name = cmd.StringArg("NAME", "", "Name of identity to use")
		getKey := getKeyClosure(cmd, allowBinaryKey)
		getValue := getValueClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			result, err := twrite.Set(tmnode(config.Node), *name, config, getKey(), getValue())
			finish(*verbose, result, err, "set")
		}
	})

	app.Command("get", "get K-V pairs", func(cmd *cli.Cmd) {
		cmd.LongDesc = "get K-V pairs\n\nIf neither -s nor -f is set, the value is base64-encoded"
		allowBinaryKey := true

		cmd.Spec = fmt.Sprintf(
			"%s %s %s %s",
			getNamespaceSpec(),
			getHeightSpec(),
			getKeySpec(allowBinaryKey),
			getEmitSpec(),
		)

		getNs := getNamespaceClosure(cmd)
		getHeight := getHeightClosure(cmd)
		getKey := getKeyClosure(cmd, allowBinaryKey)
		emit := getEmitClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			namespace := getNs(config)

			value, result, err := tool.GetNamespacedAt(
				tmnode(config.Node), namespace,
				getKey(), getHeight(),
			)
			if err == nil {
				emit(os.Stdout, value)
			}

			finish(*verbose, result, err, "get")
		}
	})

	app.Command("seq", "get the current sequence number of a namespace", func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf("%s", getNamespaceSpec())

		getNs := getNamespaceClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			ns := getNs(config)
			seq, response, err := tool.Sequence(tmnode(config.Node), ns)

			if err == nil {
				if *verbose {
					fmt.Printf("%x: %d\n", ns, seq)
				} else {
					fmt.Println(seq)
				}
			}
			finish(*verbose, response, err, "seq")
		}
	})

	app.Command("history", "get historical values for a key", func(cmd *cli.Cmd) {
		cmd.LongDesc = "get historical values for a key\n\nIf neither -s nor -f is set, the value is base64-encoded"
		allowBinaryKey := true

		cmd.Spec = fmt.Sprintf(
			"%s %s %s",
			getNamespaceSpec(),
			getKeySpec(allowBinaryKey),
			getEmitHistorySpec(),
		)

		getNs := getNamespaceClosure(cmd)
		getKey := getKeyClosure(cmd, allowBinaryKey)
		emit := getEmitHistoryClosure(cmd)

		cmd.Action = func() {
			config := getConfig()
			namespace := getNs(config)

			values, result, err := tool.HistoryNamespaced(
				tmnode(config.Node),
				namespace,
				getKey(),
				0,
				0,
			)
			if err == nil {
				emit(os.Stdout, values)
			}

			finish(*verbose, result, err, "get")
		}
	})

	app.Command("get-ns", "get all namespaces, written as base64", func(cmd *cli.Cmd) {
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
	})

	app.Command("dump", "dump all keys and values from a namespace", func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s",
			getNamespaceSpec(),
			getHeightSpec(),
			getDumpSpec(),
		)

		getHeight := getHeightClosure(cmd)
		getNamespace := getNamespaceClosure(cmd)
		dump := getDumpClosure(cmd)

		cmd.Action = func() {
			config := getConfig()

			value, result, err := tool.DumpNamespacedAt(
				tmnode(config.Node),
				getNamespace(config),
				getHeight(),
			)

			if err == nil {
				dump(os.Stdout, value)
			}
			finish(*verbose, result, err, "dump")
		}
	})

	app.Command("info", "get information about node's current status", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			info, err := tool.Info(tmnode(config.Node))
			// the whole point of this command is to get this information;
			// it makes no sense to require the verbose flag in this case
			finish(true, info, err, "info")
		}
	})

	app.Run(os.Args)
}
