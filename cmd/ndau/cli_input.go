package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/json2msgp"
	"github.com/pkg/errors"
)

func inOpt(hyphens bool, name, option string) string {
	var h string
	if hyphens {
		h = "--"
	}
	if name == "" {
		return fmt.Sprintf(
			"%s%s",
			h,
			option,
		)
	}
	return fmt.Sprintf(
		"%s%s-%s",
		h,
		strings.ToLower(name),
		option,
	)
}

func getInputSpec(name string, singleton bool) string {
	oname := name
	if singleton {
		oname = ""
	}
	return fmt.Sprintf(
		"[%s|%s|%s] (%s | %s=<PATH>) [%s=<JSON>]",
		inOpt(true, oname, "base64"),
		inOpt(true, oname, "json"),
		inOpt(true, oname, "hex"),
		strings.ToUpper(name),
		inOpt(true, oname, "file"),
		inOpt(true, oname, "json-types"),
	)
}

func getInputClosure(cmd *cli.Cmd, name string, singleton bool, verbose *bool) func() []byte {
	oname := name
	if singleton {
		oname = ""
	}
	var (
		base64In  = cmd.BoolOpt(inOpt(false, oname, "base64"), false, fmt.Sprintf("if set, interpret %s as base64-encoded", name))
		jsonIn    = cmd.BoolOpt(inOpt(false, oname, "json"), false, fmt.Sprintf("if set, interpret %s as JSON and convert to MSGP format", name))
		hexIn     = cmd.BoolOpt(inOpt(false, oname, "hex"), false, fmt.Sprintf("if set, interpret %s as hex-encoded", name))
		input     = cmd.StringArg(strings.ToUpper(name), "", "")
		inputFile = cmd.StringOpt(inOpt(false, oname, "file"), "", fmt.Sprintf("read %s from this file instead of the CLI", name))
		typesIn   = cmd.StringOpt(inOpt(false, oname, "json-types"), "", "use these type hints with json2msgp")
	)

	return func() []byte {
		var reader io.Reader
		switch {
		case input != nil && len(*input) > 0:
			if *verbose {
				fmt.Println(name, "input from cli")
			}
			reader = bytes.NewBufferString(*input)
		case inputFile != nil && len(*inputFile) > 0:
			if *verbose {
				fmt.Println(name, "input from", *inputFile)
			}
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
			if *verbose {
				fmt.Println(name, "input is b64")
			}
			out, err := base64.StdEncoding.DecodeString(string(data))
			orQuit(err)
			if *verbose {
				fmt.Printf("%s input is %d bytes long\n", name, len(out))
			}
			return out
		case hexIn != nil && *hexIn:
			if *verbose {
				fmt.Println(name, "input is hex")
			}
			out, err := hex.DecodeString(string(data))
			orQuit(err)
			if *verbose {
				fmt.Printf("%s input is %d bytes long\n", name, len(out))
			}
			return out
		case jsonIn != nil && *jsonIn:
			if *verbose {
				fmt.Println(name, "input is json -> msgp")
			}
			inbuf := bytes.NewBuffer(data)
			outbuf := &bytes.Buffer{}
			typeHints := make(map[string][]string)
			if typesIn != nil && len(*typesIn) > 0 {
				if *verbose {
					fmt.Println(name, "json type hints: ", *typesIn)
				}
				err = json.Unmarshal([]byte(*typesIn), &typeHints)
				orQuit(err)
			}
			err = json2msgp.ConvertStream(inbuf, outbuf, typeHints)
			orQuit(err)
			out := outbuf.Bytes()
			if *verbose {
				fmt.Printf("%s input is %d bytes long\n%x\n", name, len(out), out)
			}
			return out
		default:
			if *verbose {
				fmt.Println(name, "input is a string literal")
				fmt.Printf("%s input is %d bytes long\n", name, len(data))
			}
			return data
		}
	}
}
