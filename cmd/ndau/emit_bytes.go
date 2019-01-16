package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	cli "github.com/jawher/mow.cli"
)

func getEmitBytesSpec() string {
	return "[-s][-x|-r]"
}

func getEmitBytesClosure(cmd *cli.Cmd) func([]byte) {
	var (
		strip  = cmd.BoolOpt("s strip", false, "if set, do not append a newline after output")
		hexOut = cmd.BoolOpt("x hex", false, "if set, emit output as hexadecimal (default: base64)")
		rawOut = cmd.BoolOpt("r raw", false, "if set, do not encode output (implies -s) (default: base64)")
	)

	return func(bytes []byte) {
		if *rawOut {
			fmt.Print(string(bytes))
		} else if *hexOut {
			fmt.Print(hex.EncodeToString(bytes))
		} else {
			fmt.Print(base64.StdEncoding.EncodeToString(bytes))
		}
		if !*strip && !*rawOut {
			fmt.Println()
		}
	}
}
