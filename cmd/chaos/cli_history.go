package main

import (
	cli "github.com/jawher/mow.cli"
)

// getHeightSpec returns a portion of the specification string,
// specifying the point in history at which to look
func getHeightSpec() string {
	return "[-h=<HEIGHT>]"
}

// getHeightClosure sets the appropriate options for a command to get the
// appropriate moment in history
func getHeightClosure(cmd *cli.Cmd) func() uint64 {
	var (
		height = cmd.IntOpt("h height", 0, "Height of chain at which to fetch value. 0 (default) gets HEAD.")
	)

	return func() uint64 {
		var h uint64
		if height != nil {
			h = uint64(*height)
		}
		return h
	}
}
