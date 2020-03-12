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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/alexflint/go-arg"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndau/pkg/tool"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/pkg/errors"
	"github.com/savaki/jq"
)

// Tx manipulates a prepared transaction
type Tx struct{}

var _ Command = (*Tx)(nil)

// Name implements Command
func (Tx) Name() string { return "tx" }

type txargs struct {
	Name          string                 `arg:"-n" help:"with -j, name of tx to stage"`
	FromJSON      string                 `arg:"-j,--json" help:"with -n, stage a tx based on this JSON data"`
	Account       string                 `arg:"-a" help:"associate the staged tx with this account"`
	Sequence      uint                   `arg:"-s" help:"set tx sequence to this value and re-sign with account keys"`
	OverrideKey   string                 `arg:"-k,--key" help:"with -v, override this key to that value"`
	OverrideValue string                 `arg:"-v,--value" help:"with -k, override that key to this value"`
	Signatures    []signature.Signature  `arg:"separate" help:"add raw signatures to this tx"`
	SignWith      []signature.PrivateKey `arg:"separate" help:"add signatures to this tx with this private key"`
	Hash          bool                   `arg:"-h" help:"print the tx hash and return"`
	SignableBytes bool                   `arg:"-b,--signable-bytes" help:"print the base64 signable bytes of this tx and return"`
	Clear         bool                   `arg:"-C" help:"clear the staged tx"`
	Prevalidate   bool                   `arg:"-p" help:"prevalidate the tx"`
	Send          bool                   `help:"send this tx to the blockchain"`
	JQ            string                 `help:"filter output json by this jq expression"`
}

func (txargs) Description() string {
	return strings.TrimSpace(`
Manipulate a staged Tx.

-n, -j, and -a are intended to work together to construct a tx from scratch.
-a is optional, but -n and -j must be specified together if at all.
	`)
}

// Run implements Command
func (Tx) Run(argvs []string, sh *Shell) (err error) {
	args := txargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	if args.Name != "" && args.FromJSON != "" {
		if sh.Staged != nil && sh.Staged.Tx != nil {
			return errors.New("can't overwrite existing staged tx; try --clear")
		}

		if sh.Staged == nil {
			sh.Staged = &Stage{}
		}

		sh.Staged.Tx, err = ndau.TxFromName(args.Name)
		if err != nil {
			sh.Staged = nil
			return errors.Wrap(err, "getting tx from name")
		}

		err = json.Unmarshal([]byte(args.FromJSON), &sh.Staged.Tx)
		if err != nil {
			return errors.Wrap(err, "unmarshaling tx from json")
		}
	}

	if sh.Staged == nil || sh.Staged.Tx == nil {
		return errors.New("no tx currently staged")
	}

	if args.Account != "" {
		sh.Staged.Account, err = sh.Accts.Get(args.Account)
		if err != nil {
			return errors.Wrap(err, "getting account")
		}
	}

	if args.Sequence > 0 {
		err = sh.Staged.Override("sequence", args.Sequence)
		if err != nil {
			return errors.Wrap(err, "updating sequence")
		}
	}

	if args.OverrideKey != "" || args.OverrideValue != "" {
		if (args.OverrideKey == "") != (args.OverrideValue == "") {
			return errors.New("-k and -v can only be used together")
		}

		var value interface{}
		value, err = parseJSON(args.OverrideValue)
		if err != nil {
			return err
		}
		err = sh.Staged.Override(args.OverrideKey, value)
		if err != nil {
			return errors.Wrap(err, "updating "+args.OverrideKey)
		}
	}

	if len(args.Signatures) > 0 {
		err = sh.Staged.Sign(args.Signatures)
		if err != nil {
			return err
		}
	}

	if len(args.SignWith) > 0 {
		s, ok := sh.Staged.Tx.(ndau.Signable)
		if !ok {
			return fmt.Errorf("%s doesn't implement ndau.Signable", metatx.NameOf(sh.Staged.Tx))
		}
		sigs := make([]signature.Signature, 0, len(args.SignWith))
		sb := sh.Staged.Tx.SignableBytes()
		for _, pvt := range args.SignWith {
			sigs = append(sigs, pvt.Sign(sb))
		}
		s.ExtendSignatures(sigs)
		sh.Staged.Tx = s.(metatx.Transactable)
	}

	if args.Hash {
		sh.Write(metatx.Hash(sh.Staged.Tx))
		return
	}

	if args.SignableBytes {
		sh.Write(base64.StdEncoding.EncodeToString(sh.Staged.Tx.SignableBytes()))
		return
	}

	if args.Clear {
		sh.Staged = nil
		return
	}

	if args.Prevalidate {
		logger := logrus.New()
		fee, sib, _, err := tool.Prevalidate(sh.Node, sh.Staged.Tx, logger)
		if err != nil {
			return errors.Wrap(err, "prevalidating")
		}
		sh.Write("prevalidation estimates:\nfee: %s ndau\nsib: %s ndau", fee, sib)
	}

	if args.Send {
		_, err = tool.SendCommit(sh.Node, sh.Staged.Tx)
		if err != nil {
			return errors.Wrap(err, "sending to blockchain")
		}
		sh.Staged = nil
		return
	}

	if sh.Staged.Account != nil {
		sh.Staged.Account.display(sh, nil)
	}
	var data []byte
	data, err = json.MarshalIndent(sh.Staged.Tx, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling staged tx for display")
	}

	if args.JQ != "" {
		op, err := jq.Parse(args.JQ)
		if err != nil {
			return errors.Wrap(err, "parsing JQ selector")
		}
		data, err = op.Apply(data)
		if err != nil {
			return errors.Wrap(err, "applying JQ selector")
		}
	}

	sh.Write(string(data))

	return
}
