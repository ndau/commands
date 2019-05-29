package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
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
	Name          string                 `arg:"-n" help:"with -j, name of tx to stage"`
	FromJSON      string                 `arg:"-j,--json" help:"with -n, stage a tx based on this JSON data"`
	Account       string                 `arg:"-a" help:"associate the staged tx with this account"`
	Sequence      uint                   `arg:"-s" help:"set tx sequence to this value and re-sign with account keys"`
	OverrideKey   string                 `arg:"-k,--key" help:"with -v, override this key to that value"`
	OverrideValue string                 `arg:"-v,--value" help:"with -k, override that key to this value"`
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
		err = sh.Staged.Override("sequence", args.Sequence)
		if err != nil {
			return errors.Wrap(err, "updating sequence")
		}
	}

	if args.OverrideKey != "" || args.OverrideValue != "" {
		if (args.OverrideKey == "") != (args.OverrideValue == "") {
			return errors.New("-k and -v can only be used together")
		}

		// people will probably want to use bare strings for values sometimes.
		// When are they doing so?
		var autoquote bool
		if args.OverrideValue == "null" ||
			(args.OverrideValue[0] == '"' && args.OverrideValue[len(args.OverrideValue)-1] == '"') ||
			(args.OverrideValue[0] == '[' && args.OverrideValue[len(args.OverrideValue)-1] == ']') ||
			(args.OverrideValue[0] == '{' && args.OverrideValue[len(args.OverrideValue)-1] == '}') {
			autoquote = false
		} else if _, err := strconv.ParseFloat(args.OverrideValue, 64); err == nil {
			autoquote = false
		} else {
			autoquote = true
		}
		if autoquote {
			args.OverrideValue = fmt.Sprintf("\"%s\"", args.OverrideValue)
		}

		// we have to json-unmarshal the value in order to ensure that
		// we set the right datatype
		var value interface{}
		err = json.Unmarshal([]byte(args.OverrideValue), &value)
		if err != nil {
			return errors.Wrap(err, "interpreting value as JSON")
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
