package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
)

var txnames map[string]metatx.Transactable

func init() {
	// initialize txnames map
	txnames = make(map[string]metatx.Transactable)
	// add all tx full names
	for _, example := range ndau.TxIDs {
		txnames[strings.ToLower(metatx.NameOf(example))] = example
	}
	// add common abbreviations
	txnames["rfe"] = ndau.TxIDs[3]        // releasefromendowment
	txnames["claim"] = ndau.TxIDs[10]     // claimaccount
	txnames["nnr"] = ndau.TxIDs[13]       // nominatenodereward
	txnames["cvc"] = ndau.TxIDs[16]       // commandvalidatorchange
	txnames["sidechain"] = ndau.TxIDs[17] // sidechaintx
}

func knownNames() []string {
	out := make([]string, 0, len(txnames))
	for n := range txnames {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func getSignableBytes(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[-s][-x|-r] TXNAME [PATH]"

		var (
			strip  = cmd.BoolOpt("s strip", false, "if set, do not append a newline after output")
			hexOut = cmd.BoolOpt("x hex", false, "if set, emit output as hexadecimal (default: base64)")
			rawOut = cmd.BoolOpt("r raw", false, "if set, do not encode output (implies -s) (default: base64)")
			txname = cmd.StringArg("TXNAME", "", "transaction name")
			path   = cmd.StringArg("PATH", "", "if set, read tx data from this file (default: stdin)")
		)

		cmd.Action = func() {
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

			sb := tx.SignableBytes()

			if *rawOut {
				fmt.Print(string(sb))
			} else if *hexOut {
				fmt.Print(hex.EncodeToString(sb))
			} else {
				fmt.Print(base64.StdEncoding.EncodeToString(sb))
			}
			if !*strip && !*rawOut {
				fmt.Println()
			}
		}
	}
}
