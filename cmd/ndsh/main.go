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

const (
	exitAlways = iota
	exitIfErr
	exitIfNoErr
	exitNever
)

func main() {
	args := struct {
		Net     string `arg:"-N" help:"net to configure: ('main', 'test', 'dev', 'local', or a URL)"`
		Node    int    `arg:"-n" help:"node number to which to connect"`
		Verbose bool   `arg:"-v" help:"emit additional debug data"`
		Command string `arg:"-c" help:"run this command"`
		CMode   int    `arg:"-C" help:"when to exit after running a command. 0 (default): always; 1: if no err; 2: if err; 3: never"`
	}{
		Net: "mainnet",
	}

	arg.MustParse(&args)

	client, err := getClient(args.Net, args.Node)
	check(err, "setting up connection to node")

	shell := NewShell(
		args.Verbose,
		client,
		Exit{},
		Help{},
		Recover{},
		ListAccounts{},
		Verbose{},
		Net{},
		Add{},
		Watch{},
		View{},
		New{},
		RecoverKeys{},
		Claim{},
		Tx{},
		ChangeValidation{},
	)

	if args.Command != "" {
		code := 0
		err := shell.Exec(args.Command)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			code = 1
		}
		switch args.CMode {
		case exitAlways:
			os.Exit(code)
		case exitIfErr:
			if err != nil {
				os.Exit(code)
			}
		case exitIfNoErr:
			if err == nil {
				os.Exit(code)
			}
		case exitNever:
			// nothing
		default:
			os.Exit(code)
		}
	}

	shell.Run()
}
