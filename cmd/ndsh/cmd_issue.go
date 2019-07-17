package main

import (
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
)

// Issue issues ndau from the Endowment
type Issue struct{}

var _ Command = (*Issue)(nil)

// Name implements Command
func (Issue) Name() string { return "issue" }

type issueargs struct {
	Amount string `arg:"positional,required" help:"qty of ndau to issue"`
	Stage  bool   `arg:"-S" help:"stage this tx; do not send it"`
}

func (issueargs) Description() string {
	return strings.TrimSpace(`
Officially issue some ndau which has already been RFE'd.
	`)
}

// Run implements Command
func (Issue) Run(argvs []string, sh *Shell) (err error) {
	args := issueargs{}

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

	sh.VWrite("issue %s ndau", qty)

	tx := ndau.NewIssue(
		qty,
		rfemagic.Data.Sequence+1,
		rfemagic.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, nil, rfemagic)
}
