package main

import (
	cli "github.com/jawher/mow.cli"
)

// getValueSpec returns a portion of the specification string,
// specifying value setting options
func getValueSpec() string {
	return getInputSpec("value")
}

// getValueClosure sets the appropriate options for a command to get the value
// using a variety of argument styles.
func getValueClosure(cmd *cli.Cmd) func() []byte {
	return getInputClosure(cmd, "value")
}
