package main

import (
	"encoding/base64"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/chaos/ns"
	"github.com/oneiro-ndev/chaos/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
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

	app.Command("id", "manage identities", func(cmd *cli.Cmd) {
		cmd.Command("list", "list known identities", func(subcmd *cli.Cmd) {
			subcmd.Action = func() {
				config := getConfig()
				config.EmitIdentities(os.Stdout)
			}
		})

		cmd.Command("new", "create a new identity", func(subcmd *cli.Cmd) {
			subcmd.Spec = "NAME ADDR"

			var (
				name  = subcmd.StringArg("NAME", "", "Name to associate with the new identity")
				addrs = subcmd.StringArg("ADDR", "", "ndau address of account with which to pay for this id's transactions")
			)

			subcmd.Action = func() {
				addr, err := address.Validate(*addrs)
				orQuit(err)
				config := getConfig()
				err = config.CreateIdentity(*name, addr, os.Stdout)
				orQuit(errors.Wrap(err, "Failed to create identity"))
				config.Save()
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
			result, err := tool.Set(tmnode(config.Node), *name, config, getKey(), getValue())
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
