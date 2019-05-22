package main

import (
	"errors"
	"fmt"

	"github.com/alexflint/go-arg"
)

// Verbose sets or displays verbose mode
type Verbose struct{}

var _ Command = (*Verbose)(nil)

// Name implements Command
func (Verbose) Name() string { return "verbose" }

// Run implements Command
func (Verbose) Run(argvs []string, sh *Shell) (err error) {
	args := struct {
		Set   bool `help:"turn verbose mode on"`
		Unset bool `help:"turn verbose mode off"`
	}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	if args.Set && args.Unset {
		err = errors.New("cannot simultaneously set and unset verbose mode")
	} else if args.Set {
		sh.Verbose = true
		fmt.Printf("verbose mode: %t\n", sh.Verbose)
	} else if args.Unset {
		sh.Verbose = false
	} else {
		fmt.Printf("verbose mode: %t\n", sh.Verbose)
	}

	return
}
