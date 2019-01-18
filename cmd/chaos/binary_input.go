package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/json2msgp"
	"github.com/pkg/errors"
)

func getInputSpec(name string) string {
	return fmt.Sprintf(
		"[-b|-j|-x] (%s | --%s-file=<PATH>)",
		strings.ToUpper(name),
		strings.ToLower(name),
	)
}

func getInputClosure(cmd *cli.Cmd, name string) func() []byte {
	var (
		base64In  = cmd.BoolOpt("b base64", false, "if set, interpret input as base64-encoded")
		jsonIn    = cmd.BoolOpt("j json", false, "if set, interpret input as JSON and convert to MSGP format")
		hexIn     = cmd.BoolOpt("x hex", false, "if set, interpret input as hex-encoded")
		input     = cmd.StringArg(strings.ToUpper(name), "", fmt.Sprintf("%s input", name))
		inputFile = cmd.StringOpt(fmt.Sprintf("%s-file", name), "", "read input from this file instead of the CLI")
	)

	return func() []byte {
		var reader io.Reader
		switch {
		case input != nil && len(*input) > 0:
			reader = bytes.NewBufferString(*input)
		case inputFile != nil && len(*inputFile) > 0:
			file, err := os.Open(*inputFile)
			orQuit(err)
			defer file.Close()
			reader = file
		default:
			orQuit(errors.New("no input provided"))
		}

		data, err := ioutil.ReadAll(reader)
		orQuit(err)

		switch {
		case base64In != nil && *base64In:
			out, err := base64.StdEncoding.DecodeString(string(data))
			orQuit(err)
			return out
		case hexIn != nil && *hexIn:
			out, err := hex.DecodeString(string(data))
			orQuit(err)
			return out
		case jsonIn != nil && *jsonIn:
			inbuf := bytes.NewBuffer(data)
			outbuf := &bytes.Buffer{}
			err = json2msgp.ConvertStream(inbuf, outbuf)
			orQuit(err)
			return outbuf.Bytes()
		default:
			return data
		}
	}
}
