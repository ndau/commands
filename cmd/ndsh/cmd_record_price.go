package main

import (
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
)

// RecordPrice issues released funds
type RecordPrice struct{}

var _ Command = (*RecordPrice)(nil)

// Name implements Command
func (RecordPrice) Name() string { return "record-price" }

type recordpriceargs struct {
	Dollars string `arg:"positional,required" help:"record this quantity of dollars as the current price"`
	Stage   bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (recordpriceargs) Description() string {
	return strings.TrimSpace(`
Send a random number from which a node reward winner is derived.
	`)
}

// Run implements Command
func (RecordPrice) Run(argvs []string, sh *Shell) (err error) {
	args := recordpriceargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var magic *Account
	magic, err = sh.SAAcct(sv.RecordPriceAddressName, sv.RecordPriceValidationPrivateName)
	if err != nil {
		return
	}

	var nanocents pricecurve.Nanocent
	nanocents, err = pricecurve.ParseDollars(args.Dollars)
	if err != nil {
		return
	}

	sh.VWrite("recording current price: %d nanocents", nanocents)

	tx := ndau.NewRecordPrice(
		nanocents,
		magic.Data.Sequence+1,
		magic.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, nil, magic)
}
