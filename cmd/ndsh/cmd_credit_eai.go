package main

import (
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/pkg/errors"
)

// CreditEAI an account
type CreditEAI struct{}

var _ Command = (*CreditEAI)(nil)

// Name implements Command
func (CreditEAI) Name() string { return "credit-eai" }

type crediteaiargs struct {
	Account string `arg:"positional" help:"account whose delegates to credit eai for"`
	Update  bool   `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage   bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (crediteaiargs) Description() string {
	return strings.TrimSpace(`
Credit EAI to accounts delegated to an acccount.
	`)
}

// Run implements Command
func (CreditEAI) Run(argvs []string, sh *Shell) (err error) {
	args := crediteaiargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var acct *Account
	acct, err = sh.Accts.Get(args.Account)
	if err != nil {
		return errors.Wrap(err, "account")
	}

	if acct.Data == nil || args.Update {
		err = acct.Update(sh, sh.Write)
		if IsAccountDoesNotExist(err) {
			err = nil
		}
		if err != nil {
			return
		}
	}

	sh.VWrite("credit eai for %s", acct.Address)

	tx := ndau.NewCreditEAI(
		acct.Address,
		acct.Data.Sequence+1,
		acct.PrivateValidationKeys...,
	)
	return sh.Dispatch(args.Stage, tx, acct, nil)
}
