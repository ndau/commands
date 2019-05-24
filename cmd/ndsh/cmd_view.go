package main

import (
	"encoding/json"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/pkg/errors"
)

// View leaves the shell
type View struct{}

var _ Command = (*View)(nil)

// Name implements Command
func (View) Name() string { return "view" }

type viewargs struct {
	Account string `arg:"positional,required" help:"view this account"`
	Update  bool   `arg:"-u" help:"update this account from the blockchain before viewing"`
	// TODO:
	// JQ string `help:"filter output json by this jq expression"`
}

func (viewargs) Description() string {
	return strings.TrimSpace(`
View an account's data.

By default, this operates only on cached data. To get current data from the
blockchain, use the --update flag.
	`)
}

// Run implements Command
func (View) Run(argvs []string, sh *Shell) (err error) {
	args := viewargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var acct *Account
	acct, err = sh.accts.Get(args.Account)
	if err != nil {
		return
	}

	acct.display(sh, sh.accts.Reverse()[acct])

	if args.Update {
		if sh.Verbose {
			sh.Write("communicating with blockchain...")
		}
		err = acct.Update(sh, sh.Write)
		if err != nil {
			return
		}
	}

	jsdata, err := json.MarshalIndent(acct.Data, "", "  ")
	if err != nil {
		err = errors.Wrap(err, "marshalling account data to json")
		return
	}

	sh.Write(string(jsdata))
	return
}
