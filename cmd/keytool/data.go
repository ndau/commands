package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"

	cli "github.com/jawher/mow.cli"
)

func getDataSpec(allowStdin bool) string {
	if allowStdin {
		return "(DATA | --file=<path> | --stdin)  [-x|-b]"
	}
	return "(DATA | --file=<path>)  [-x|-b]"
}

func getDataClosure(cmd *cli.Cmd, allowStdin bool) func() []byte {
	var (
		texti  = cmd.StringArg("DATA", "", "data: if neither x nor b is set, it is taken as raw bytes")
		filei  = cmd.StringOpt("f file", "", "path to file containing applicable data")
		hexi   = cmd.BoolOpt("x hex", false, "interpret data as hex-encoded")
		b64    = cmd.BoolOpt("b b64", false, "interpret data as b64-encoded")
		stdini *bool
	)

	if allowStdin {
		stdini = cmd.BoolOpt("S stdin", false, "read data from stdin")
	}

	return func() []byte {
		var data []byte
		switch {
		case texti != nil && len(*texti) > 0:
			data = []byte(*texti)
		case filei != nil && len(*filei) > 0:
			var err error
			data, err = ioutil.ReadFile(*filei)
			check(err)
		case stdini != nil && *stdini:
			in := bufio.NewScanner(os.Stdin)
			if !in.Scan() {
				check(errors.New("stdin selected but empty"))
			}
			check(in.Err())
			data = []byte(in.Text())
		default:
			check(errors.New("no data provided; should be unreachable"))
		}

		switch {
		case hexi != nil && *hexi:
			ks, err := hex.DecodeString(string(data))
			check(err)
			data = []byte(ks)
		case b64 != nil && *b64:
			ks, err := base64.StdEncoding.DecodeString(string(data))
			check(err)
			data = []byte(ks)
		}

		return data
	}
}
