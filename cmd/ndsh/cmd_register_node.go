package main

import (
	"encoding/base64"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/pkg/errors"
)

// RegisterNode an account
type RegisterNode struct{}

var _ Command = (*RegisterNode)(nil)

// Name implements Command
func (RegisterNode) Name() string { return "register-node" }

type rnargs struct {
	DistributionScript string `arg:"positional,required" help:"base64 of node distribution script"`
	Account            string `arg:"positional" help:"account to register as node"`
	Update             bool   `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage              bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (rnargs) Description() string {
	return strings.TrimSpace(`
Register an account as a node.
	`)
}

// Run implements Command
func (RegisterNode) Run(argvs []string, sh *Shell) (err error) {
	args := rnargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	distScript, err := base64.StdEncoding.DecodeString(args.DistributionScript)
	if err != nil {
		return errors.Wrap(err, "distribution script")
	}

	var acct *Account
	acct, err = sh.Accts.Get(args.Account)
	if err != nil {
		return errors.Wrap(err, "account")
	}

	if acct.OwnershipPublic == nil {
		return errors.New("ownership public key unknown")
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

	sh.VWrite("registering %s as node and script %s", acct.Address, args.DistributionScript)

	tx := ndau.NewRegisterNode(
		acct.Address,
		distScript,
		*acct.OwnershipPublic,
		acct.Data.Sequence+1,
		acct.PrivateValidationKeys...,
	)
	return sh.Dispatch(args.Stage, tx, acct, nil)
}
