package main

import (
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

// Watch adds foreign accounts, sometimes with nicknames
type Watch struct{}

var _ Command = (*Watch)(nil)

// Name implements Command
func (Watch) Name() string { return "watch" }

type watchargs struct {
	Address   address.Address `arg:"positional,required" help:"watch this account"`
	Nicknames []string        `arg:"-n,separate" help:"short nicknames which can refer to this account."`
}

func (watchargs) Description() string {
	return strings.TrimSpace(`
Watch accounts for which you do not possess the private keys.

This is most useful to declare short nicknames for destination accounts.
	`)
}

// Run implements Command
func (Watch) Run(argvs []string, sh *Shell) (err error) {
	args := watchargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	a := &Account{
		Address: args.Address,
	}
	err = a.Update(sh, sh.Write)
	if err != nil {
		return
	}

	sh.accts.Add(a, args.Nicknames...)

	return
}
