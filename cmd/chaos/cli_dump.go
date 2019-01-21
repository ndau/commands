package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/chaos/query"
)

func getDumpSpec() string {
	return "[-s | -S]"
}

func getDumpClosure(cmd *cli.Cmd) func(io.Writer, *query.DumpNamespaceResponse) error {
	var (
		asString   = cmd.BoolOpt("s string", false, "Interpret keys and values as strings")
		onlyString = cmd.BoolOpt("S only-string", false, "Emit only keys and values for which both can be interpreted as valid unicode. Implies -s.")
	)

	if *onlyString {
		*asString = true
	}

	return func(w io.Writer, dnr *query.DumpNamespaceResponse) error {
		for _, kv := range dnr.Data {
			key := kv.Key
			value := kv.Value

			if *onlyString {
				if !utf8.ValidString(string(key)) || !utf8.ValidString(string(value)) {
					continue
				}
			}
			if *asString {
				fmt.Fprintf(
					w, "%s=%s\n",
					strconv.QuoteToASCII(string(key)),
					strconv.QuoteToASCII(string(value)),
				)
			} else {
				fmt.Fprintf(
					w, "%s %s\n",
					base64.StdEncoding.EncodeToString(key),
					base64.StdEncoding.EncodeToString(value),
				)
			}
		}
		return nil
	}
}
