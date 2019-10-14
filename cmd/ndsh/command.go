package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"os"

	"github.com/alexflint/go-arg"
)

// A Command acts like a sub-program: it can specify its own argument parser,
// and can succeed or fail, displaying arbitrary messages as a result, without
// exiting the shell.
type Command interface {
	// Name is the name by which this command is called.
	//
	// If more than one space-separated word is present, it is treated as a synonym.
	Name() string

	// Run invokes this command.
	//
	// Much like standard commands on the CLI, the run string is the entire
	// string typed by the user, so it will always begin with the name of
	// the command. Lexing is handled by the shell, but argument parsing is
	// the responsibility of the individual command.
	//
	// Unlike typical CLI commands, commands within ndsh have mutable access
	// to typed global state. They may also return human-readable errors.
	//
	// Long-running commands are encouraged to background themselves by doing
	// most of their work in a separate goroutine and returning immediately.
	// If they choose to do so, they should coordinate by adding to the
	// shell.Running WaitGroup before launching their goroutine and
	// defer shell.Running.Done() immediately after launching it.
	//
	// All commands taking more than trivial time, or which perform network IO,
	// should periodically attempt to read from shell.Stop; if such a read
	// produces a result, the command should shut itself down immediately.
	Run([]string, *Shell) error
}

// ParseInto parses the argvs into the destination data
//
// This leverages alexflint/go-arg, so the dest struct can be constructed
// the same way. However, the only toplevel functions in that library
// always take the argument list from os.Args, which isn't appropriate;
// this function just lets you put in the appropriate argument strings.
func ParseInto(argvs []string, dest ...interface{}) error {
	p, err := arg.NewParser(arg.Config{Program: argvs[0]}, dest...)
	if err != nil {
		return err
	}
	err = p.Parse(argvs[1:])
	if err == arg.ErrHelp {
		p.WriteHelp(os.Stdout)
	}
	return err
}
