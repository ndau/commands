package main

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/words"
	"github.com/pkg/errors"
)

const defaultPathFmt = "/44'/20036'/100/%d"
const defaultPathIdx = 1

var defaultPath = fmt.Sprintf(defaultPathFmt, defaultPathIdx)

// New creates a new account
type New struct{}

var _ Command = (*New)(nil)

// Name implements Command
func (New) Name() string { return "new" }

type newargs struct {
	ShareSeed string   `arg:"-S,--share-seed" help:"share the seed phrase with this account"`
	SeedSize  uint     `arg:"--seed-size" help:"num bytes of random seed used to generate seed phrase"`
	Path      string   `arg:"-p" help:"create account with this derivation path"`
	PathIdx   uint     `arg:"-i,--path-idx" help:"use this value as the account index"`
	Lang      string   `arg:"-l" help:"seed phrase language"`
	Kind      string   `arg:"-k" help:"kind of account"`
	Nicknames []string `arg:"-n,--nickname,separate" help:"short nicknames which can refer to this account"`
}

func (newargs) Description() string {
	return strings.TrimSpace(`
Create a new account.

If shareseed is set, it will share a seed with that account. Otherwise a new
seed phrase will be generated and displayed. This phrase will only ever be
displayed at this time, so be sure to copy it to somewhere secure!

If path is set, it will be used. If unset, a sensible default will be used:
if shareseed is set, the path will increment the highest path sharing that
seed; otherwise, it will use the constant /44'/20036'/100/1. Note: when setting
the path within the shell, watch out for single quotes: they must be escaped or
enclosed within surrounding quotes.

Note: this does not touch the blockchain, it only constructs the account in
internal data.
	`)
}

// Run implements Command
func (New) Run(argvs []string, sh *Shell) (err error) {
	args := newargs{
		Path:     defaultPath,
		PathIdx:  defaultPathIdx,
		Lang:     "en",
		Kind:     string(address.KindUser),
		SeedSize: 16,
	}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	var root *key.ExtendedKey
	root, err = sh.getRoot(args.ShareSeed, &args.Path, &args.PathIdx, args.SeedSize, args.Lang)
	if err != nil {
		return
	}

	kind, err := address.ParseKind(args.Kind)
	if err != nil {
		return errors.Wrap(err, "invalid kind")
	}

	var acct Account
	acct, err = newAccountFromRoot(root, args.Path, kind)
	acct.AcctIdx = int(args.PathIdx)
	if err != nil {
		return errors.Wrap(err, "constructing account")
	}
	acct.display(sh, args.Nicknames)

	sh.Accts.Add(&acct, args.Nicknames...)

	return
}

func (sh *Shell) getRoot(sharedname string, path *string, pathidx *uint, seedsize uint, lang string) (root *key.ExtendedKey, err error) {
	if sharedname != "" {
		var shared *Account
		shared, err = sh.Accts.Get(sharedname)
		if err != nil {
			return
		}
		if shared.root == nil {
			err = fmt.Errorf("%s root is unknown; cannot share it", sharedname)
			return
		}
		root = shared.root

		// we might also have to update the path
		if *path == defaultPath && shared.Path != "" {
			if *pathidx == defaultPathIdx {
				// we need to detect an appropriate path index
				// what's the highest path we know of which shares this seed?
				var idx uint
				_, err = fmt.Sscanf(shared.Path, defaultPathFmt, pathidx)
				if err != nil {
					err = errors.Wrap(err, "getting idx of path of shared account")
					return
				}

				for _, acct := range sh.Accts.accts {
					if acct.root == shared.root && acct.Path != "" {
						_, err = fmt.Sscanf(acct.Path, defaultPathFmt, &idx)
						if err != nil {
							sh.Write("%s: %s", acct.Address, err.Error())
							continue
						}
						if idx > *pathidx {
							*pathidx = idx
						}
					}
				}

				*pathidx++
			}
			// else a path index was picked by the user
			*path = fmt.Sprintf(defaultPathFmt, *pathidx)
		}
	} else {
		wordsseed := make([]byte, seedsize)
		_, err = rand.Read(wordsseed)
		if err != nil {
			err = errors.Wrap(err, "generating bytes for seed")
			return
		}

		root, err = key.NewMaster(wordsseed)
		if err != nil {
			err = errors.Wrap(err, "generating root key from seed")
			return
		}

		var seedphrase []string
		seedphrase, err = words.FromBytes(lang, wordsseed)
		if err != nil {
			err = errors.Wrap(err, "converting seed phrase bytes to words")
			return
		}

		sh.Write("seed phrase:\n%s", strings.Join(seedphrase, " "))
	}
	return
}
