package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// Claim claims an account, assigning its first validation keys and script
type Claim struct{}

var _ Command = (*Claim)(nil)

const (
	secureKeypath = "/44'/20036'/100/10000'/%d'/%d"
	walletKeypath = "/44'/20036'/100/10000/%d/%d"
)

// Name implements Command
func (Claim) Name() string { return "claim set-validation" }

type claimargs struct {
	Account          string   `arg:"positional" help:"account to claim"`
	NumKeys          uint     `arg:"-n,--num-keys" help:"number of validation keys to set"`
	Paths            []string `arg:"-p,separate" help:"use these keypaths"`
	ValidationScript string   `arg:"-s,--script" help:"set this validation script (base64)"`
	WalletCompat     bool     `arg:"-C,--wallet-compat" help:"if set, generate keypaths the way the wallet does"`
	Update           bool     `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage            bool     `arg:"-S" help:"stage this tx; do not send it"`
}

func (claimargs) Description() string {
	return strings.TrimSpace(`
Claim an account.

All paths specified will be used. If paths are not set, appropriate paths will
be generated.

By default, secure keypaths will be generated. However, for compatibility with
the ndau wallet, insecure keypaths can be used.
	`)
}

// Run implements Command
func (Claim) Run(argvs []string, sh *Shell) (err error) {
	args := claimargs{
		NumKeys: 1,
	}

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

	if acct.root == nil {
		return errors.New("account root unknown; can't derive keys")
	}
	if acct.OwnershipPublic == nil {
		return errors.New("account public ownership key unknown; can't generate claim tx")
	}
	if acct.OwnershipPrivate == nil {
		return errors.New("account private ownership key unknown; can't sign claim tx")
	}

	var validationScript []byte
	if args.ValidationScript != "" {
		validationScript, err = base64.StdEncoding.DecodeString(args.ValidationScript)
		if err != nil {
			return errors.Wrap(err, "decoding validation script")
		}
	}

	var (
		pvts []signature.PrivateKey
		pubs []signature.PublicKey
	)

	for _, path := range args.Paths {
		if sh.Verbose {
			sh.Write("generating key at %s...", path)
		}
		pubs, pvts, err = derive(acct.root, path, pubs, pvts)
		if err != nil {
			return
		}
		args.NumKeys--
	}

	var keypath string
	if args.WalletCompat {
		keypath = walletKeypath
	} else {
		keypath = secureKeypath
	}

	for ; args.NumKeys > 0; args.NumKeys-- {
		acct.HighKeyIdx++
		path := fmt.Sprintf(keypath, acct.AcctIdx, acct.HighKeyIdx)
		if sh.Verbose {
			sh.Write("generating key at %s...", path)
		}
		pubs, pvts, err = derive(acct.root, path, pubs, pvts)
		if err != nil {
			return
		}
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

	tx := ndau.NewSetValidation(
		acct.Address,
		*acct.OwnershipPublic,
		pubs,
		validationScript,
		acct.Data.Sequence+1,
		*acct.OwnershipPrivate,
	)

	err = sh.Dispatch(args.Stage, tx, acct, nil)
	if err != nil {
		return
	}

	acct.PrivateValidationKeys = pvts
	return
}

func derive(root *key.ExtendedKey, path string, pubs []signature.PublicKey, pvts []signature.PrivateKey) ([]signature.PublicKey, []signature.PrivateKey, error) {
	child, err := root.DeriveFrom("/", path)
	if err != nil {
		return pubs, pvts, errors.Wrap(err, "deriving child key")
	}

	pvt, err := child.SPrivKey()
	if err != nil {
		return pubs, pvts, errors.Wrap(err, "converting child pvt key to ndau fmt")
	}
	pub, err := child.SPubKey()
	if err != nil {
		return pubs, pvts, errors.Wrap(err, "converting child pub key to ndau fmt")
	}

	pubs = append(pubs, *pub)
	pvts = append(pvts, *pvt)
	return pubs, pvts, nil
}
