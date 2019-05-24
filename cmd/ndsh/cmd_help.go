package main

import (
	"fmt"
	"sort"

	"github.com/alexflint/go-arg"
)

// Help provides help about general usage and specific commands
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
		names := make([]string, 0)
		for _, command := range sh.Commands {
			name := command.Name()
			if _, ok := knownNames[name]; ok {
				// skip it
			} else {
				knownNames[name] = struct{}{}
				names = append(names, name)
			}
		}
		sort.Strings(names)

		sh.WriteBatch(func(print func(format string, context ...interface{})) {
			print("Ndau Shell: common ndau operations on in-memory data")
			print("")
			print("Known commands:")
			for _, name := range names {
				print("  %s", name)
			}
			print("")
			print("(`help command` to get detail help on that particular command")
		})
	}
	return
}
