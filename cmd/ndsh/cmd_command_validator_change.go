package main

import (
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
)

// CommandValidatorChange changes the validation power of a node
type CommandValidatorChange struct{}

var _ Command = (*CommandValidatorChange)(nil)

// Name implements Command
func (CommandValidatorChange) Name() string { return "command-validator-change cvc" }

type cvcargs struct {
	Power int64  `arg:"positional,required" help:"power to assign to this node"`
	Node  string `arg:"positional" help:"node to assign power to"`
	Stage bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (cvcargs) Description() string {
	return strings.TrimSpace(`
Command a Validator power change.

By setting power to 0, turn a validator into a verifier.
By setting power to non-0, turn a verifier into a validator.
	`)
}

// Run implements Command
func (CommandValidatorChange) Run(argvs []string, sh *Shell) (err error) {
	args := cvcargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var magic *Account
	magic, err = sh.SAAcct(sv.CommandValidatorChangeAddressName, sv.CommandValidatorChangeValidationPrivateName)
	if err != nil {
		return
	}

	node, err := sh.Accts.Get(args.Node)
	if err != nil {
		return
	}

	sh.VWrite("set validator power of %s to %d", node.Address, args.Power)

	tx := ndau.NewCommandValidatorChange(
		node.Address,
		args.Power,
		magic.Data.Sequence+1,
		magic.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, nil, magic)
}
