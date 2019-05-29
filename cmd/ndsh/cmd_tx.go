package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alexflint/go-arg"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// Tx claims an account, assigning its first validation keys and script
type Tx struct{}

var _ Command = (*Tx)(nil)

// Name implements Command
func (Tx) Name() string { return "tx" }

type txargs struct {
	Name          string                 `arg:"-n" help:"name of tx to stage (with -j)"`
	FromJSON      string                 `arg:"-j" help:"stage a tx based on this JSON data (with -n)"`
	Account       string                 `arg:"-a" help:"associate the staged tx with this account"`
	Sequence      uint                   `arg:"-s" help:"set tx sequence to this value and re-sign with account keys"`
	Signatures    []signature.Signature  `arg:"separate" help:"add raw signatures to this tx"`
	SignWith      []signature.PrivateKey `arg:"separate" help:"add signatures to this tx with this private key"`
	SignableBytes bool                   `arg:"-b" help:"print the base64 signable bytes of this tx and return"`
	Clear         bool                   `arg:"-C" help:"clear the staged tx"`
	Send          bool                   `help:"send this tx to the blockchain"`
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
		if sh.Staged != nil || sh.Staged.Tx != nil {
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
		sh.Staged.Account, err = sh.accts.Get(args.Account)
		if err != nil {
			return errors.Wrap(err, "getting account")
		}
	}

	if args.Sequence > 0 {
		// this is a bit of a nasty hack, but we can't get around it: we know
		// that every tx has a Sequence field, and we'd like to manipulate it.
		// Go, however, doesn't know that every tx has a Sequence field.
		// It therefore won't let us do that.
		// What we can do is marshal it to JSON, unmarshal it into a dict,
		// set the sequence, and unmarshal.
		var data []byte
		var jsdata map[string]interface{}
		data, err = json.Marshal(sh.Staged.Tx)
		if err != nil {
			return errors.Wrap(err, "marshaling staged tx for sequence edit")
		}
		err = json.Unmarshal(data, &jsdata)
		if err != nil {
			return errors.Wrap(err, "unmarshaling tx into map for sequence edit")
		}

		// update sequence
		jsdata["sequence"] = args.Sequence

		// clear existing signatures: they'll be invalid now
		jsdata["signature"] = nil
		jsdata["signatures"] = nil

		// clear out the tx object
		txid, err := metatx.TxIDOf(sh.Staged.Tx, ndau.TxIDs)
		if err != nil {
			return errors.Wrap(err, "getting txid")
		}
		sh.Staged.Tx = metatx.Clone(ndau.TxIDs[txid])

		data, err = json.Marshal(jsdata)
		if err != nil {
			return errors.Wrap(err, "marshaling sequence edited tx")
		}
		err = json.Unmarshal(data, &sh.Staged.Tx)
		if err != nil {
			return errors.Wrap(err, "unmarshaling sequence edited tx")
		}

		// re-sign with associated account
		if sh.Staged.Account != nil {
			err = sh.Staged.Sign(nil)
			if err != nil {
				return err
			}
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

	if args.SignableBytes {
		sh.Write(base64.StdEncoding.EncodeToString(sh.Staged.Tx.SignableBytes()))
		return
	}

	if args.Clear {
		sh.Staged = nil
		return
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
	sh.Write(string(data))

	return
}
