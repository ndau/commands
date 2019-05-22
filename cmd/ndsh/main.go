package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
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
	args := struct {
		Neturl  string `arg:"-N" help:"net to configure: ('main', 'test', 'dev', 'local', or a URL)"`
		Node    int    `arg:"-n" help:"node number to which to connect"`
		Verbose bool   `arg:"-v" help:"emit additional debug data"`
	}{
		Neturl: "mainnet",
	}

	arg.MustParse(&args)

	shell := NewShell(
		args.Verbose,
		getClient(args.Neturl, args.Node),
		Exit{},
		Help{},
		Recover{},
		ListAccounts{},
		Verbose{},
	)
	shell.Run()
}
