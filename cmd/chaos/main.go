package main

import (
	"os"

	cli "github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("chaos", "interact with the chaos chain")

	app.Spec = "[-v]"

	var (
		verbose = app.BoolOpt("v verbose", false, "Emit detailed results from the chaos chain if set")
	)

	app.Command("conf", "perform initial configuration", getCmdConf(verbose))
	app.Command("conf-path", "show location of config file", getCmdConfPath())
	app.Command("import-assc", "import an assc.toml file", getCmdImportAssc())
	app.Command("id", "manage identities", getCmdID())
	app.Command("set", "set K-V pairs", getCmdSet(verbose))
	app.Command("get", "get K-V pairs", getCmdGet(verbose))
	app.Command("seq", "get the current sequence number of a namespace", getCmdSeq(verbose))
	app.Command("history", "get historical values for a key", getCmdHistory(verbose))
	app.Command("get-ns", "get all namespaces, written as base64", getCmdGetNS(verbose))
	app.Command("dump", "dump all keys and values from a namespace", getCmdDump(verbose))
	app.Command("info", "get information about node's current status", getCmdInfo())

	app.Run(os.Args)
}
