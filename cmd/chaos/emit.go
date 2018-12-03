package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/chaos/query"
)

// getEmitSpec returns a portion of the specification string,
// specifying value emission options
func getEmitSpec() string {
	return "[(-s | -t)... | -f=<FILE>]"
}

// getEmitClosure returns a closure which emits the specified bytes per
// the options specified on the command line
func getEmitClosure(cmd *cli.Cmd) func(io.Writer, []byte) error {
	var (
		asString = cmd.BoolOpt("s string", false, "Interpret the returned value as a utf-8 string")
		trim     = cmd.BoolOpt("t trim", false, "Do not add a newline after the value")
		file     = cmd.StringOpt("f file", "", "Write the returned value to a file named by FILE. Implies -t. Does not encode the data as base64.")
	)

	return func(w io.Writer, value []byte) error {
		// if file is set, just emit the output to the file
		if len(*file) > 0 {
			// default permissions: u=rw;go-
			return ioutil.WriteFile(*file, value, 0600)
		}

		if *asString {
			// interpret as string and emit
			fmt.Fprint(w, string(value))
		} else {
			// interpret as bytes; emit base64 encoding
			b64 := base64.StdEncoding.EncodeToString(value)
			fmt.Fprint(w, b64)
		}

		// newline if not trim
		if !*trim {
			fmt.Fprint(w, "\n")
		}

		return nil
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
