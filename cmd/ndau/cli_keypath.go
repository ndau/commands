package main

import cli "github.com/jawher/mow.cli"

func getKeypathSpec(optional bool) string {
	if optional {
		return "[-P=<keypath>]"
	}
	return "KEYPATH"
}

func getKeypathClosure(cmd *cli.Cmd, optional bool) func() string {
	var keypath *string

	if optional {
		keypath = cmd.StringOpt("P keypath", "", "derivation path of key")
	} else {
		keypath = cmd.StringArg("KEYPATH", "", "derivation path of key")
	}

	return func() string {
		if keypath == nil {
			return ""
		}
		return *keypath
	}
}
