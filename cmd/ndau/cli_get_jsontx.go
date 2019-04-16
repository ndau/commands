package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	cli "github.com/jawher/mow.cli"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
)

func getJSONTXSpec() string {
	return "TXNAME [PATH]"
}

func getJSONTXClosure(cmd *cli.Cmd) func() metatx.Transactable {
	var (
		txname = cmd.StringArg("TXNAME", "", "transaction name")
		path   = cmd.StringArg("PATH", "", "if set, read tx data from this file (default: stdin)")
	)

	return func() metatx.Transactable {
		// prepare tx
		tx, err := ndau.TxFromName(*txname)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			for _, tx := range ndau.KnownTxNames() {
				fmt.Fprintln(os.Stderr, "  ", tx)
			}
			os.Exit(1)

		}

		// prepare input
		var reader *bufio.Reader
		if path == nil || *path == "" {
			reader = bufio.NewReader(os.Stdin)
		} else {
			f, err := os.Open(*path)
			orQuit(err)
			reader = bufio.NewReader(f)
		}
		bytes, err := ioutil.ReadAll(reader)
		orQuit(err)

		// load tx
		err = json.Unmarshal(bytes, tx)
		orQuit(err)

		return tx
	}
}
