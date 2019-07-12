package main

import (
	"strings"

	"github.com/oneiro-ndev/ndaumath/pkg/address"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
)

// ReleaseFromEndowment releases ndau from the endowment
type ReleaseFromEndowment struct{}

var _ Command = (*ReleaseFromEndowment)(nil)

// Name implements Command
func (ReleaseFromEndowment) Name() string { return "release-from-endowment rfe" }

type rfeargs struct {
	Amount  string `arg:"positional,required" help:"qty of ndau to rfe"`
	Account string `arg:"positional" help:"account to rfe into"`
	Stage   bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (rfeargs) Description() string {
	return strings.TrimSpace(`
Release funds from the endowment into an account.
	`)
}

// Run implements Command
func (ReleaseFromEndowment) Run(argvs []string, sh *Shell) (err error) {
	args := rfeargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var rfemagic *Account
	rfemagic, err = sh.SAAcct(sv.ReleaseFromEndowmentAddressName, sv.ReleaseFromEndowmentValidationPrivateName)
	if err != nil {
		return
	}

	var qty math.Ndau
	qty, err = math.ParseNdau(args.Amount)
	if err != nil {
		return
	}

	var acct *Account
	var addr *address.Address
	addr, acct, err = sh.AddressOf(args.Account)
	if err != nil {
		return
	}

	sh.VWrite("rfe %s ndau to %s", qty, *addr)

	tx := ndau.NewReleaseFromEndowment(
		*addr,
		qty,
		rfemagic.Data.Sequence+1,
		rfemagic.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, acct, rfemagic)
}
