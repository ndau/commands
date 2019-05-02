package main

import (
	"errors"
	"strings"
)

// Exit leaves the shell
type Exit struct{}

var _ Command = (*Exit)(nil)

// Name implements Command
func (Exit) Name() string { return "exit quit" }

// Run implements Command
func (Exit) Run(args []string, sh *Shell) error {
	var err error
	if len(args) > 1 {
		err = errors.New(strings.Join(args[1:], " "))
	}
	sh.Exit(err)
	return err
}
