package main

import (
	"strings"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/words"
	"github.com/pkg/errors"
)

// Add adds accounts and/or nicknames
type Add struct{}

var _ Command = (*Add)(nil)

// Name implements Command
func (Add) Name() string { return "add" }

type runargs struct {
	SeedFrom   string   `arg:"-r,--seed-from" help:"use the same seed as this account"`
	SeedPhrase string   `arg:"-S,--seed-phrase" help:"seed phrase generating the required seed"`
	Path       string   `arg:"-p" help:"create account with this derivation path"`
	Nicknames  []string `arg:"-n,separate" help:"short nicknames which can refer to this account. Only applied if exactly one account was recovered"`
	Lang       string   `arg:"-l" help:"recovery phrase language"`
	Kind       string   `arg:"-k" help:"kind of account"`
}

func (runargs) Description() string {
	return strings.TrimSpace(`
Add accounts by path, or add nicknames to an account.

Example: add an account with an explicit derivation path, skipping discovery:

	add -S "eye eye eye eye eye eye eye eye eye eye eye eye" -p "/1'/2'/3/4"

Example: add an explicit derivation path from a seed in another account:

	add -r 5sx -p "/1'/2'/4/8"

Example: add some nicknames to an existing account:

	add -r 5sx -n myaccount -n bank

Note the quotes around the paths: single quotes have semantic value in paths,
but they're _also_ shell-style quote chars, so they need to be escaped or enclosed.
	`)
}

// prerequisite: r.Path is set
func (r runargs) acct(sh *Shell) (*Account, error) {
	if r.SeedFrom != "" && r.SeedPhrase != "" {
		return nil, errors.New("cannot set both `-r` and `-S`")
	}
	if r.SeedFrom != "" {
		return sh.accts.Get(r.SeedFrom)
	}
	if r.SeedPhrase != "" {
		seedwords := strings.Split(r.SeedPhrase, " ")

		if len(seedwords) != 12 {
			sh.Write("WARN: ndau seed phrases are typically 12 words long, but you provided %d\n", len(r.SeedPhrase))
		}
		for idx := range seedwords {
			seedwords[idx] = strings.ToLower(seedwords[idx])
		}
		seed, err := words.ToBytes(r.Lang, seedwords)
		if err != nil {
			return nil, err
		}

		root, err := key.NewMaster(seed)
		if err != nil {
			return nil, err
		}

		kind, err := address.ParseKind(r.Kind)
		if err != nil {
			return nil, err
		}

		acctstream := make(chan Account, 1)
		var wg sync.WaitGroup
		wg.Add(1)

		go sh.tryAccount(root, 0, r.Path, kind, acctstream, &wg)
		wg.Wait()
		close(acctstream)
		acct, ok := <-acctstream

		// that channel may or may not have returned a real value
		// if not, we want to reconstruct the account and move forward anyway
		if !ok {
			acct, err = NewAccount(seed, r.Path, kind)
			if err != nil {
				return nil, errors.Wrap(err, "constructing non-blockchain acct")
			}
		}

		// if we already know about this account, return the one we know of
		// instead of the same one at a different pointer
		sacct, err := sh.accts.Get(acct.Address.String())
		if err == nil {
			return sacct, nil
		}

		return &acct, nil
	}

	return nil, errors.New("must specify `-r` or `-S` to get a root phrase")
}

// Run implements Command
func (Add) Run(argvs []string, sh *Shell) (err error) {
	args := runargs{
		Lang: "en",
		Kind: string(address.KindUser),
	}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	if sh.Verbose && args.Path != "" {
		sh.Write("check quoting of -p arg: %s", args.Path)
	}

	var acct *Account
	if args.Path == "" && args.SeedFrom == "" {
		err = errors.New("must set at least one of `-p` or `-r`")
	} else if args.Path != "" && args.SeedFrom != "" {
		// special case: add an account with a new derivation path (-p)
		// from the same seed which derived an existing account (-r)
		var existing *Account
		existing, err = sh.accts.Get(args.SeedFrom)
		if err != nil {
			return
		}
		var kind byte
		kind, err = address.ParseKind(args.Kind)
		if err != nil {
			return err
		}
		var racct Account
		racct, err = newAccountFromRoot(existing.root, args.Path, kind)
		acct = &racct
	} else if args.Path != "" {
		// if we've set a path, we expect to create an account
		acct, err = args.acct(sh)
	} else if args.SeedFrom != "" {
		acct, err = sh.accts.Get(args.SeedFrom)
	}
	if err != nil {
		return
	}

	sh.accts.Add(acct, args.Nicknames...)

	return
}
