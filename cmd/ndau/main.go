package main

import (
	"os"

	cli "github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("ndau", "interact with the ndau chain")

	app.Spec = "[-v][-k][-j [-c]]"

	var (
		verbose  = app.BoolOpt("v verbose", false, "emit detailed results from the ndau chain if set")
		keys     = app.IntOpt("k keys", -1, "bitset of keys to use in signature. when negative, use all available")
		emitJSON = app.BoolOpt("j json", false, "emit tx as JSON instead of sending to node")
		compact  = app.BoolOpt("c compact", false, "emit compact JSON (default: pretty)")
	)

	app.Command("conf", "perform initial configuration", getConf(verbose))
	app.Command("conf-path", "show location of config file", confPath)
	app.Command("account", "manage accounts", getAccount(verbose, keys, emitJSON, compact))
	app.Command("transfer", "transfer ndau from one account to another", getTransfer(verbose, keys, emitJSON, compact))
	app.Command("transfer-lock", "transfer ndau from one account to a new account and lock the destination", getTransferAndLock(verbose, keys, emitJSON, compact))
	app.Command("rfe", "release ndau from the endowment", getRfe(verbose, keys, emitJSON, compact))
	app.Command("nnr", "nominate node reward", getNNR(verbose, keys, emitJSON, compact))
	app.Command("info", "get information about node's current status", getInfo(verbose))
	app.Command("cvc", "send a command validator change", getCVC(verbose, keys, emitJSON, compact))
	app.Command("server", "create server for API endpoint calls", server)
	app.Command("summary", "emit summary information about the ndau chain", getSummary(verbose))
	app.Command("version", "emit version information and quit", getVersion(verbose))

	app.Run(os.Args)
}
