package main

import (
	"fmt"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
)

func bail(err string, context ...interface{}) {
	if !strings.HasSuffix(err, "\n") {
		err = err + "\n"
	}
	fmt.Fprintf(os.Stderr, err, context...)
	os.Exit(1)
}

func check(err error, context string, more ...interface{}) {
	if err != nil {
		bail(fmt.Sprintf("%s: %s", context, err.Error()), more...)
	}
}

func main() {
	app := cli.App("ndsh", "the ndau shell")

	var (
		neturl  = app.StringOpt("N net", "mainnet", "net to configure: ('main', 'test', 'dev', 'local', or a url)")
		node    = app.IntOpt("n node", 0, "node number to which to connect")
		verbose = app.BoolOpt("v verbose", false, "emit additional debug data")
	)

	app.Action = func() {
		shell := NewShell(
			*verbose,
			getClient(*neturl, *node),
			Exit{},
		)
		shell.Run()
	}

	app.Run(os.Args)
}
