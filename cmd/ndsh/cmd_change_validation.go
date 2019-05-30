package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// ChangeValidation changes an account's validation
type ChangeValidation struct{}

var _ Command = (*ChangeValidation)(nil)

// Name implements Command
func (ChangeValidation) Name() string { return "change-validation" }

type cvargs struct {
	Account          string   `arg:"positional" help:"account to claim"`
	NumKeys          uint     `arg:"-n,--add-keys" help:"number of validation keys to add"`
	Paths            []string `arg:"-p,separate" help:"use these keypaths"`
	RemoveKeyIdx     []uint   `arg:"-r,--remove-key,separate" help:"remove the existing validation key at this index (0-based)"`
	ValidationScript string   `arg:"-s,--script" help:"set this validation script (base64)"`
	WalletCompat     bool     `arg:"-C,--wallet-compat" help:"if set, generate keypaths the way the wallet does"`
	Update           bool     `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage            bool     `arg:"-S" help:"stage this tx; do not send it"`
}

func (cvargs) Description() string {
	return strings.TrimSpace(`
Change an account's validation.

All paths specified will be used. If paths are not set, appropriate paths will
be generated.

By default, secure keypaths will be generated. However, for compatibility with
the ndau wallet, insecure keypaths can be used.

By default all existing keys are retained. Keys may be removed by specifying
the 0-based index of the key to remove.
	`)
}

// Run implements Command
func (ChangeValidation) Run(argvs []string, sh *Shell) (err error) {
	args := cvargs{}

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
	if len(acct.PrivateValidationKeys) == 0 {
		return errors.New("0 validation keys known; consider recover-keys")
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

	var validationScript []byte
	switch args.ValidationScript {
	case "":
		validationScript = acct.Data.ValidationScript
	case "null", "nil":
		// validationScript is nil
	default:
		validationScript, err = base64.StdEncoding.DecodeString(args.ValidationScript)
		if err != nil {
			return errors.Wrap(err, "decoding validation script")
		}
	}

	var (
		pvts    = make([]signature.PrivateKey, 0, len(args.Paths)+int(args.NumKeys)+len(acct.PrivateValidationKeys))
		pubs    = make([]signature.PublicKey, 0, len(args.Paths)+int(args.NumKeys)+len(acct.Data.ValidationKeys))
		removed = make([]signature.PublicKey, 0, len(args.RemoveKeyIdx))
	)

	// filter existing validation keys into those removed and those retained
	sh.VWrite("filtering validation keys: %d exist", len(acct.Data.ValidationKeys))
	for idx, pub := range acct.Data.ValidationKeys {
		removethis := false
		for _, r := range args.RemoveKeyIdx {
			if idx == int(r) {
				removethis = true
				break
			}
		}
		if removethis {
			removed = append(removed, pub)
		} else {
			pubs = append(pubs, pub)
		}
	}
	sh.VWrite("  %d removed; %d retained", len(removed), len(pubs))

	// generate list of retained private keys
	sh.VWrite("filtering private keys; %d known", len(acct.PrivateValidationKeys))
	for _, pvt := range acct.PrivateValidationKeys {
		match := false
		for _, pub := range removed {
			if signature.Match(pub, pvt) {
				match = true
				break
			}
		}
		if !match {
			pvts = append(pvts, pvt)
		}
	}
	sh.VWrite("  %d retained", len(pvts))

	for _, path := range args.Paths {
		sh.VWrite("generating key at %s...", path)
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
		sh.VWrite("generating key at %s...", path)
		pubs, pvts, err = derive(acct.root, path, pubs, pvts)
		if err != nil {
			return
		}
	}

	tx := ndau.NewChangeValidation(
		acct.Address,
		pubs,
		validationScript,
		acct.Data.Sequence+1,
		acct.PrivateValidationKeys...,
	)

	return sh.Dispatch(args.Stage, tx, acct, nil)
}
