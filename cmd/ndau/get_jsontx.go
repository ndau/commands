package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
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
		example, ok := txnames[strings.ToLower(*txname)]
		if !ok {
			fmt.Fprintln(os.Stderr, "Unknown transaction: ", *txname)
			fmt.Println("Known transactions:")
			for _, tx := range knownNames() {
				fmt.Println("  ", tx)
			}
			os.Exit(1)
		}

		tx := metatx.Clone(example)

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
