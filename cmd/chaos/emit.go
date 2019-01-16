package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"unicode/utf8"

	"github.com/tinylib/msgp/msgp"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/chaos/query"
	"github.com/pkg/errors"
)

func getEmitSpec() string {
	return "[-S][-m|-r|-s|-x]"
}

func getEmitClosure(cmd *cli.Cmd) func([]byte) {
	var (
		strip     = cmd.BoolOpt("S strip", false, "if set, do not append a newline after output")
		msgpOut   = cmd.BoolOpt("m msgp", false, "if set, convert msgp output to json")
		rawOut    = cmd.BoolOpt("r raw", false, "if set, do not encode output (implies -S) (default: base64)")
		stringOut = cmd.BoolOpt("s string", false, "if set, interpret output as utf-8")
		hexOut    = cmd.BoolOpt("x hex", false, "if set, emit output as hexadecimal (default: base64)")
	)

	return func(output []byte) {
		switch {
		case *msgpOut:
			_, err := msgp.CopyToJSON(os.Stdout, bytes.NewBuffer(output))
			orQuit(err)
		case *rawOut, *stringOut:
			fmt.Print(string(output))
			if *stringOut && !utf8.Valid(output) {
				orQuit(errors.New(
					"output was not a string: " + base64.StdEncoding.EncodeToString(output),
				))
			}
			fmt.Print(string(output))
		case *hexOut:
			fmt.Print(hex.EncodeToString(output))
		default:
			fmt.Print(base64.StdEncoding.EncodeToString(output))
		}

		if !(*strip || *rawOut) {
			fmt.Println()
		}
	}
}

// getEmitHistorySpec returns a portion of the specification string,
// specifying value emission options for multiple values
func getEmitHistorySpec() string {
	return "[-s | -f=<FILE>]"
}

// getEmitHistoryClosure returns a closure which emits each of the specified
// byte slices per the options specified on the command line
func getEmitHistoryClosure(cmd *cli.Cmd) func(io.Writer, *query.KeyHistoryResponse) error {
	var (
		asString = cmd.BoolOpt("s string", false, "Interpret the value as a utf-8 string")
		file     = cmd.StringOpt("f file", "", "Write the values to a series of files named FILE.1, FILE.2, etc. Does not encode the data as base64.")
	)

	return func(w io.Writer, khr *query.KeyHistoryResponse) error {
		var err error
		// if file is set, just emit the output to the file
		if len(*file) > 0 {
			for _, hv := range khr.History {
				fn := *file + "." + string(hv.Height)
				// default permissions: u=rw;go-
				err = ioutil.WriteFile(fn, hv.Value, 0600)
				if err != nil {
					return err
				}
			}
		}

		for _, hv := range khr.History {
			fmt.Fprintf(w, "Height %d:\n", hv.Height)
			if *asString {
				// interpret as string and emit
				fmt.Fprint(w, string(hv.Value))
			} else {
				// interpret as bytes; emit base64 encoding
				b64 := base64.StdEncoding.EncodeToString(hv.Value)
				fmt.Fprint(w, b64)
			}
			fmt.Fprint(w, "\n\n")
		}

		return nil
	}
}
