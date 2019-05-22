package main

import (
	"strings"

	"github.com/alexflint/go-arg"
)

// ListAccounts lists all known accounts
type ListAccounts struct{}

var _ Command = (*ListAccounts)(nil)

// Name implements Command
func (ListAccounts) Name() string { return "accounts" }

// Run implements Command
func (ListAccounts) Run(argvs []string, sh *Shell) (err error) {
	args := struct {
	}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	for acct, nicknames := range sh.accts.Reverse() {
		sh.Write("%s: %s", acct.Address, strings.Join(nicknames, " "))
	}
	return
}
