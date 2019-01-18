package main

import (
	cli "github.com/jawher/mow.cli"
)

// getKeySpec returns a portion of the specification string,
// specifying key setting options
func getKeySpec() string {
	return getInputSpec("key")
}

// getKeyClosure sets the appropriate options for a command to get the key
// using a variety of argument styles.
func getKeyClosure(cmd *cli.Cmd) func() []byte {
	return getInputClosure(cmd, "key")
}
