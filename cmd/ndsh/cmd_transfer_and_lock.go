package main

import (
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// TransferAndLock claims an account, assigning its first validation keys and script
type TransferAndLock struct{}

var _ Command = (*TransferAndLock)(nil)

// Name implements Command
func (TransferAndLock) Name() string { return "transfer-lock tnl" }

type transferlockargs struct {
	Qty      string        `arg:"positional,required" help:"qty to transfer in ndau"`
	From     string        `arg:"positional,required" help:"account to transfer from. Use \"\" for inference"`
	To       string        `arg:"positional,required" help:"account to transfer to. Any full address is valid even if not otherwise known."`
	Duration math.Duration `arg:"positional,required" help:"Duration of the lock"`
	Stage    bool          `arg:"-S" help:"stage this tx; do not send it"`
}

func (transferlockargs) Description() string {
	return strings.TrimSpace(`
Transfer ndau from one account to another, locking the recipient.
	`)
}

// Run implements Command
func (TransferAndLock) Run(argvs []string, sh *Shell) (err error) {
	args := transferlockargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var from, to *Account
	from, err = sh.Accts.Get(args.From)
	if err != nil {
		return
	}

	var toaddr address.Address
	to, err = sh.Accts.Get(args.To)
	switch {
	case err == nil:
		toaddr = to.Address
	case IsNoMatch(err):
		toaddr, err = address.Validate(args.To)
		if err != nil {
			return errors.Wrap(err, "To must be known address substring, nickname, or full address")
		}
	case err != nil:
		return
	}

	var qty math.Ndau
	qty, err = math.ParseNdau(args.Qty)
	if err != nil {
		return
	}

	sh.VWrite("transfering %s ndau (%d napu) from %s to %s", qty, qty, from.Address, toaddr)

	tx := ndau.NewTransferAndLock(
		from.Address,
		toaddr,
		qty,
		args.Duration,
		from.Data.Sequence+1,
		from.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, from, nil)
}
