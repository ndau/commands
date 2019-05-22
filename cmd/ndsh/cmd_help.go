package main

import (
	"fmt"

	"github.com/alexflint/go-arg"
)

// Help leaves the shell
type Help struct{}

var _ Command = (*Help)(nil)

// Name implements Command
func (Help) Name() string { return "help ?" }

// Run implements Command
func (Help) Run(argvs []string, sh *Shell) (err error) {
	args := struct {
		Command string `arg:"positional" help:"display detailed help about this command"`
	}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	if args.Command != "" {
		sh.Exec(fmt.Sprintf("%s -h", args.Command))
	} else {
		knownNames := make(map[string]struct{})
		fmt.Println("Ndau Shell: common ndau operations on in-memory data")
		fmt.Println()
		fmt.Println("Known commands:")
		for _, command := range sh.Commands {
			name := command.Name()
			if _, ok := knownNames[name]; ok {
				// skip it
			} else {
				knownNames[name] = struct{}{}
				fmt.Printf("  %s\n", name)
			}
		}
		fmt.Println()
		fmt.Println("(`help command` to get detail help on that particular command")
	}
	return
}
