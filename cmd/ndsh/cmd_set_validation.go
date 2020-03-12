package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndaumath/pkg/key"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// SetValidation assigns an account's first validation keys and script
type SetValidation struct{}

var _ Command = (*SetValidation)(nil)

const (
	secureKeypath = "/44'/20036'/100/10000'/%d'/%d"
	walletKeypath = "/44'/20036'/100/10000/%d/%d"
)

var autorecoverValidation []byte

func init() {
	var err error
	autorecoverValidation, err = base64.StdEncoding.DecodeString("oAARiKABAiCI")
	if err != nil {
		panic(err)
	}
}

const recoveryServicePath = "/tx/submit/setvalidation"

// Name implements Command
func (SetValidation) Name() string { return "set-validation" }

type validationargs struct {
	Account          string   `arg:"positional" help:"account to modify"`
	NumKeys          uint     `arg:"-n,--num-keys" help:"number of validation keys to set"`
	Paths            []string `arg:"-p,separate" help:"use these keypaths"`
	ValidationScript string   `arg:"-s,--script" help:"set this validation script (base64)"`
	WalletCompat     bool     `arg:"-C,--wallet-compat" help:"if set, generate keypaths the way the wallet does"`
	Update           bool     `arg:"-u" help:"update this account from the blockchain before creating tx"`
	Stage            bool     `arg:"-S" help:"stage this tx; do not send it"`
	Autorecover      bool     `help:"send validation tx to recovery service if account looks like it might be subscribed. When true, -S may not work if recovery is attempted. Disable with --autorecover=false"`
}

func (validationargs) Description() string {
	return strings.TrimSpace(`
Set validation rules for an account.

All paths specified will be used. If paths are not set, appropriate paths will
be generated.

By default, secure keypaths will be generated. However, for compatibility with
the ndau wallet, insecure keypaths can be used.
	`)
}

// Run implements Command
func (SetValidation) Run(argvs []string, sh *Shell) (err error) {
	args := validationargs{
		NumKeys:     1,
		Autorecover: true,
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
		return errors.New("account public ownership key unknown; can't generate set-validation tx")
	}
	if acct.OwnershipPrivate == nil {
		return errors.New("account private ownership key unknown; can't sign set-validation tx")
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

	if len(acct.Data.ValidationKeys) == 2 &&
		bytes.Equal(acct.Data.ValidationScript, autorecoverValidation) {
		sh.VWrite(
			"acct might be subscribed to recovery service; autorecover: %v",
			args.Autorecover,
		)
		if args.Autorecover {
			if RecoveryURL == nil {
				return errors.New("no known recovery service for this net")
			}

			txjson, err := json.Marshal(tx)
			if err != nil {
				return errors.Wrap(err, "marshaling json for recovery service")
			}
			buf := bytes.NewBuffer(txjson)

			// recovery service needs its path
			url := RecoveryURL.String() + recoveryServicePath
			resp, err := http.Post(url, "application/json", buf)
			if err != nil {
				return errors.Wrap(err, "sending request to recovery service")
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				if err != nil {
					return fmt.Errorf(
						"recovery service returned code %s and resp body could not be read: %s",
						resp.Status,
						err,
					)
				}
				return fmt.Errorf(
					"recovery service returned code %s:\n%s",
					resp.Status,
					string(body),
				)
			}

			sh.Write("recovery service request returned success; watch blockchain for updates")
			sh.VWrite(string(body))
			return nil
		}
	}

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
