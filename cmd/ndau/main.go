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
	app.Command("currency-seats", "list all currency seats on the blockchain", getCurrencySeats(verbose))
	app.Command("show-delegates", "emit information about the chain's delegates", getDelegates(verbose))
	app.Command("transfer", "transfer ndau from one account to another", getTransfer(verbose, keys, emitJSON, compact))
	app.Command("transfer-lock", "transfer ndau from one account to a new account and lock the destination", getTransferAndLock(verbose, keys, emitJSON, compact))
	app.Command("rfe", "release ndau from the endowment", getRfe(verbose, keys, emitJSON, compact))
	app.Command("issue", "issue ndau that have been rfe'd", getIssue(verbose, keys, emitJSON, compact))
	app.Command("nnr", "nominate node reward", getNNR(verbose, keys, emitJSON, compact))
	app.Command("cvc", "send a command validator change", getCVC(verbose, keys, emitJSON, compact))
	app.Command("record-price", "record the current market price of ndau", getRecordPrice(verbose, keys, emitJSON, compact))
	app.Command("record-endowment-nav", "record the net asset value of the endowment", getRecordEndowmentNAV(verbose, keys, emitJSON, compact))
	app.Command("sysvar", "get and set system variables", getSysvar(verbose, keys, emitJSON, compact))
	app.Command("summary", "emit summary information about the ndau chain", getSummary(verbose))
	app.Command("current-sib sib", "emit information about the current SIB", getSIB(verbose))
	app.Command("info", "get information about node's current status", getInfo(verbose))
	app.Command("version", "emit version information and quit", getVersion(verbose))
	app.Command("signable-bytes", "emit the signable bytes of the input tx", getSignableBytes(verbose))
	app.Command("send", "send a pre-prepared transaction", getSendJSON(verbose))

	app.Run(os.Args)
}
