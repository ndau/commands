package main

import (
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// ChangeRecourse claims an account, assigning its first validation keys and script
type ChangeRecourse struct{}

var _ Command = (*ChangeRecourse)(nil)

// Name implements Command
func (ChangeRecourse) Name() string { return "change-recourse" }

type changerecourseargs struct {
	Period  math.Duration `arg:"positional,required" help:"new recourse period"`
	Account string        `arg:"positional" help:"account to change recourse period of"`
	Update  bool          `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage   bool          `arg:"-S" help:"stage this tx; do not send it"`
}

func (changerecourseargs) Description() string {
	return strings.TrimSpace(`
Change an account's recourse period.

Note that this will not take effect until the old recourse period has expired.
	`)
}

// Run implements Command
func (ChangeRecourse) Run(argvs []string, sh *Shell) (err error) {
	args := changerecourseargs{}

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
		return
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

	sh.VWrite("changing recourse period of %s to %s", acct.Address, args.Period)

	tx := ndau.NewChangeRecoursePeriod(
		acct.Address,
		args.Period,
		acct.Data.Sequence+1,
		acct.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, acct, nil)
}
