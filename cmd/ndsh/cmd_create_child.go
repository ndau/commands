package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// CreateChild creates a child account, assigning its initial attributes
type CreateChild struct{}

var _ Command = (*CreateChild)(nil)

// Name implements Command
func (CreateChild) Name() string { return "create-child child" }

type createchildargs struct {
	Parent           string        `arg:"positional" help:"parent account"`
	WalletCompat     bool          `arg:"-C,--wallet-compat" help:"if set, generate keypaths the way the wallet does"`
	DelegationNode   string        `arg:"-d,--delegation-node" help:"delegate the child to this node"`
	PathIdx          uint          `arg:"-i,--path-idx" help:"use this value as the account index"`
	KeyPaths         []string      `arg:"-k,--key-path,separate" help:"use these keypaths"`
	Kind             string        `arg:"-K" help:"kind of account"`
	Lang             string        `arg:"-l" help:"seed phrase language"`
	Nicknames        []string      `arg:"-n,--nickname,separate" help:"short nicknames which can refer to this account"`
	NumKeys          uint          `arg:"-N,--num-keys" help:"number of validation keys to set"`
	RecoursePeriod   math.Duration `arg:"-p,--recourse-period" help:"recourse period for child account"`
	Path             string        `arg:"-P" help:"create account with this derivation path"`
	ValidationScript string        `arg:"-s,--script" help:"set this validation script (base64)"`
	Stage            bool          `arg:"-S" help:"stage this tx; do not send it"`
	ShareSeed        bool          `arg:"--share-seed" help:"share the seed phrase with the parent account"`
	SeedSize         uint          `arg:"--seed-size" help:"num bytes of random seed used to generate seed phrase"`
}

func (createchildargs) Description() string {
	return strings.TrimSpace(`
Create a child account.

All keypaths specified will be used. If paths are not set, appropriate paths will
be generated.

By default, secure keypaths will be generated. However, for compatibility with
the ndau wallet, insecure keypaths can be used.
	`)
}

// Run implements Command
func (CreateChild) Run(argvs []string, sh *Shell) (err error) {
	args := createchildargs{
		NumKeys:        1,
		Path:           defaultPath,
		PathIdx:        defaultPathIdx,
		Lang:           "en",
		Kind:           string(address.KindUser),
		SeedSize:       16,
		RecoursePeriod: 1 * math.Hour,
	}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	// parent
	var parent *Account
	parent, err = sh.Accts.Get(args.Parent)
	if err != nil {
		return
	}

	if parent.root == nil {
		return errors.New("account root unknown; can't derive keys")
	}

	// child
	var share string
	if args.ShareSeed {
		share = parent.Address.String()
	}
	var root *key.ExtendedKey
	root, err = sh.getRoot(share, &args.Path, &args.PathIdx, args.SeedSize, args.Lang)
	if err != nil {
		return
	}

	kind, err := address.ParseKind(args.Kind)
	if err != nil {
		return errors.Wrap(err, "invalid kind")
	}

	var child Account
	child, err = newAccountFromRoot(root, args.Path, kind)
	child.AcctIdx = int(args.PathIdx)
	if err != nil {
		return errors.Wrap(err, "constructing child account")
	}

	// validation script
	var validationScript []byte
	if args.ValidationScript != "" {
		validationScript, err = base64.StdEncoding.DecodeString(args.ValidationScript)
		if err != nil {
			return errors.Wrap(err, "decoding validation script")
		}
	}

	// validation keys
	var (
		pvts []signature.PrivateKey
		pubs []signature.PublicKey
	)

	for _, path := range args.KeyPaths {
		if sh.Verbose {
			sh.Write("generating key at %s...", path)
		}
		pubs, pvts, err = derive(parent.root, path, pubs, pvts)
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
		child.HighKeyIdx++
		path := fmt.Sprintf(keypath, child.AcctIdx, child.HighKeyIdx)
		if sh.Verbose {
			sh.Write("generating key at %s...", path)
		}
		pubs, pvts, err = derive(child.root, path, pubs, pvts)
		if err != nil {
			return
		}
	}

	// delegation node
	if args.DelegationNode == "" {
		args.DelegationNode = args.Parent
	}
	var delegateAddr *address.Address
	delegateAddr, _, err = sh.AddressOf(args.DelegationNode)
	if err != nil {
		return errors.Wrap(err, "getting delegation node")
	}

	// ensure we have a parent
	if parent.Data == nil {
		err = parent.Update(sh, sh.Write)
		if err != nil {
			return errors.Wrap(err, "updating parent")
		}
	}

	// create and send tx
	tx := ndau.NewCreateChildAccount(
		parent.Address,
		child.Address,
		*child.OwnershipPublic,
		child.OwnershipPrivate.Sign([]byte(child.Address.String())),
		args.RecoursePeriod,
		pubs,
		validationScript,
		*delegateAddr,
		parent.Data.Sequence+1,
		parent.PrivateValidationKeys...,
	)

	err = sh.Dispatch(args.Stage, tx, parent, nil)
	if err != nil {
		return
	}

	err = child.Update(sh, sh.Write)
	if args.Stage && IsAccountDoesNotExist(err) {
		err = nil
	}
	if err != nil {
		return
	}

	child.display(sh, args.Nicknames)
	sh.Accts.Add(&child, args.Nicknames...)

	return nil
}
