package main

import (
	"encoding/base64"
	"io/ioutil"

	"github.com/pkg/errors"

	cli "github.com/jawher/mow.cli"
)

// getValueSpec returns a portion of the specification string,
// specifying value setting options
func getValueSpec() string {
	return "[-v=<VAL> | -V=<VAL> | --value-file=<VF> | --binary-value-file=<VF>]"
}

// getValueClosure sets the appropriate options for a command to get the value
// using a variety of argument styles.
func getValueClosure(cmd *cli.Cmd) func() []byte {
	var (
		value = cmd.StringOpt("v value", "", "Interpret VAL as text on the command line")
		vf    = cmd.StringOpt("value-file", "", "Interpret VF as the path to a utf-8 encoded file containing the desired value")
		bval  = cmd.StringOpt("V binary-value", "", "Interpret VAL as base64-encoded data on the command line")
		bvf   = cmd.StringOpt("binary-value-file", "", "Interpret VF as the path to a binary file containing the desired value")
	)

	return func() []byte {
		if value != nil && *value != "" {
			// must be valid utf-8
			bytes := []byte(*value)
			validateBytes(bytes)
			return bytes
		}
		if vf != nil && *vf != "" {
			k, err := ioutil.ReadFile(*vf)
			orQuit(errors.Wrap(err, "Bad value file"))
			// must be valid utf-8
			validateBytes(k)
			return k
		}

		if bval != nil && *bval != "" {
			b, err := base64.StdEncoding.DecodeString(*bval)
			orQuit(errors.Wrap(err, "Bad value string"))
			return b
		}
		if bvf != nil && *bvf != "" {
			k, err := ioutil.ReadFile(*bvf)
			orQuit(errors.Wrap(err, "Bad value file"))
			return k
		}

		// if no value option is set, return an empty slice
		return make([]byte, 0)
	}
}
