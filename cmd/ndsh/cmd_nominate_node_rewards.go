package main

import (
	"crypto/rand"
	"encoding/binary"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// NominateNodeRewards nominates a node account to receive rewards
type NominateNodeRewards struct{}

var _ Command = (*NominateNodeRewards)(nil)

// Name implements Command
func (NominateNodeRewards) Name() string { return "nominate-node-rewards nnr" }

type nnrargs struct {
	Random int64 `help:"use this number instead of generating one randomly"`
	Stage  bool  `arg:"-S" help:"stage this tx; do not send it"`
}

func (nnrargs) Description() string {
	return strings.TrimSpace(`
Send a random number from which a node reward winner is derived.
	`)
}

// Run implements Command
func (NominateNodeRewards) Run(argvs []string, sh *Shell) (err error) {
	args := nnrargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var magic *Account
	magic, err = sh.SAAcct(sv.NominateNodeRewardAddressName, sv.NominateNodeRewardValidationPrivateName)
	if err != nil {
		return
	}

	if args.Random == 0 {
		buf := make([]byte, 8)
		_, err = rand.Read(buf)
		if err != nil {
			return errors.Wrap(err, "getting random data")
		}
		args.Random = int64(binary.BigEndian.Uint64(buf))
	}

	tx := ndau.NewNominateNodeReward(
		args.Random,
		magic.Data.Sequence+1,
		magic.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, nil, magic)
}
