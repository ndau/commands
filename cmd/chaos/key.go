package main

import (
	"encoding/base64"
	"io/ioutil"

	"github.com/pkg/errors"

	cli "github.com/jawher/mow.cli"
)

// getKeySpec returns a portion of the specification string,
// specifying key setting options
func getKeySpec(allowBinary bool) string {
	if allowBinary {
		return "(-k=<KEY> | -K=<KEY> | --key-file=<KF> | --binary-key-file=<KF>)"
	}
	return "(-k=<KEY> | --key-file=<KF>)"
}

// getKeyClosure sets the appropriate options for a command to get the key
// using a variety of argument styles.
func getKeyClosure(cmd *cli.Cmd, allowBinary bool) func() []byte {
	var (
		key = cmd.StringOpt("k key", "", "Interpret KEY as text on the command line")
		kf  = cmd.StringOpt("key-file", "", "Interpret KF as the path to a utf-8 encoded file containing the desired key")
	)

	var bkey, bkf *string = nil, nil
	if allowBinary {
		bkey = cmd.StringOpt("K binary-key", "", "Interpret KEY as base64-encoded data on the command line")
		bkf = cmd.StringOpt("binary-key-file", "", "Interpret KF as the path to a binary file containing the desired key")
	}

	return func() []byte {
		if key != nil && *key != "" {
			// key must be valid utf-8
			bytes := []byte(*key)
			validateBytes(bytes)
			return bytes
		}
		if kf != nil && *kf != "" {
			k, err := ioutil.ReadFile(*kf)
			orQuit(errors.Wrap(err, "bad key file"))
			// key must be valid utf-8
			validateBytes(k)
			return k
		}
		if allowBinary {
			if bkey != nil && *bkey != "" {
				b, err := base64.StdEncoding.DecodeString(*bkey)
				orQuit(errors.Wrap(err, "bad binary key"))
				return b
			}
			if bkf != nil && *bkf != "" {
				k, err := ioutil.ReadFile(*bkf)
				orQuit(errors.Wrap(err, "bad binary key file"))
				return k
			}
		}

		// we shouldn't get here; one of the above key flags should have been set
		orQuit(errors.New("No key set"))
		panic("unreachable")
	}
}
