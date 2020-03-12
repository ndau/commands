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
	"fmt"
	"strings"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/key"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/ndau/ndaumath/pkg/words"
	"github.com/pkg/errors"
)

var accountPatterns = []string{
	"/%d",
	"/44'/20036'/100/%d",
}

// Recover an account from its seed phrase
type Recover struct{}

var _ Command = (*Recover)(nil)

// Name implements Command
func (Recover) Name() string { return "recover" }

// Run implements Command
func (Recover) Run(argvs []string, sh *Shell) (err error) {
	args := struct {
		SeedPhrase  []string             `arg:"positional" help:"seed phrase from which to recover this account"`
		Root        signature.PrivateKey `help:"recover from this root key instead of a seed phrase"`
		Nicknames   []string             `arg:"-n,separate" help:"short nicknames which can refer to this account. Only applied if exactly one account was recovered"`
		Lang        string               `arg:"-l" help:"recovery phrase language"`
		Persistence int                  `help:"number of non-accounts to discover before deciding there are no more in a derivation style"`
		Kind        string               `arg:"-k" help:"kind of account"`
	}{
		Lang:        "en",
		Persistence: 50,
		Kind:        string(address.KindUser),
	}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	if (len(args.SeedPhrase) == 0) == args.Root.IsZero() {
		return errors.New("must specify seed phrase or root")
	}

	kind, err := address.ParseKind(args.Kind)
	if err != nil {
		return err
	}

	var root *key.ExtendedKey
	if len(args.SeedPhrase) > 0 {
		if len(args.SeedPhrase) != 12 {
			sh.Write("WARN: ndau seed phrases are typically 12 words long, but you provided %d\n", len(args.SeedPhrase))
		}
		for idx := range args.SeedPhrase {
			(args.SeedPhrase)[idx] = strings.ToLower((args.SeedPhrase)[idx])
		}

		seed, err := words.ToBytes(args.Lang, args.SeedPhrase)
		if err != nil {
			return err
		}

		root, err = key.NewMaster(seed)
		if err != nil {
			return errors.Wrap(err, "generating root key")
		}
	}
	if !args.Root.IsZero() {
		root = &key.ExtendedKey{}
		err = root.FromSignatureKey(&args.Root)
		if err != nil {
			return errors.Wrap(err, "converting root key")
		}
	}

	sh.Write("Communicating with blockchain...")
	// TODO: add some kind of progress bar?
	// https://github.com/vbauerster/mpb seems good for this usage

	accountsStream := make(chan Account, 0)
	var wg sync.WaitGroup
	accounts := make([]Account, 0)

	// for each account pattern, keep trying variations until persist failures
	for _, pattern := range accountPatterns {
		patstream := make(chan Account, 0)
		patmutex := sync.Mutex{}
		patidx := 0
		var patwg sync.WaitGroup
		// for every success, pass it to the outer stream, but also
		// try a new path as well
		go func(pattern string) {
			for acct := range patstream {
				accountsStream <- acct

				patmutex.Lock()
				path := fmt.Sprintf(pattern, patidx)

				// note: we don't increment the wgs here for the new goroutine
				// that's the responsibility of the old goroutine
				go sh.tryAccount(root, patidx, path, kind, patstream, &wg, &patwg)

				patidx++
				patmutex.Unlock()
			}
		}(pattern)

		// now generate the initial set of attempts for this pattern
		patmutex.Lock()
		for patidx = 0; patidx < args.Persistence; patidx++ {
			wg.Add(1)
			patwg.Add(1)

			path := fmt.Sprintf(pattern, patidx)
			go sh.tryAccount(root, patidx, path, kind, patstream, &wg, &patwg)
		}
		patmutex.Unlock()

		// close the pattern stream when we exhaust the pattern possibilities
		go func(pattern string) {
			patwg.Wait()
			close(patstream)
		}(pattern)
	}

	// collect all results from all patterns into an array
	go func() {
		for acct := range accountsStream {
			accounts = append(accounts, acct)
		}
	}()

	wg.Wait()

	sh.Write("Discovered %d accounts:", len(accounts))
	for _, acct := range accounts {
		// we're about to add by pointer.
		// Go will reuse the same struct for each acct, though.
		// We need to make a copy in order to ensure that all the pointers are
		// distinct.
		acctCopy := acct
		sh.Write("  %s (%s)\n", acct.Address, acct.Path)
		sh.Accts.Add(&acctCopy)
	}
	// add nicknames if we've recovered exactly one account
	if len(accounts) == 1 && len(args.Nicknames) > 0 {
		sh.Write("Adding nicknames to %s: %s", accounts[0].Address, strings.Join(args.Nicknames, ", "))
		sh.Accts.Add(&accounts[0], args.Nicknames...)
	}

	return err
}

// try getting an account from the blockchain. If it exists, construct an appropriate
// struct and pass it along the channel. Discard any errors.
//
// each invocation of this function should be run within a new goroutine
//
// Does _not_ attempt to discover any private keys
func (sh *Shell) tryAccount(
	root *key.ExtendedKey,
	acctidx int,
	path string,
	kind byte,
	out chan<- Account,
	wgs ...*sync.WaitGroup,
) {
	defer func() {
		for _, wg := range wgs {
			wg.Done()
		}
	}()

	sh.WriteBatch(func(print func(format string, args ...interface{})) {
		if sh.Verbose {
			print("tryAccount(%s, %s)", path, string(kind))
		}

		acct, err := newAccountFromRoot(root, path, kind)
		acct.AcctIdx = acctidx
		if err != nil {
			if sh.Verbose {
				print("    newaccount: %s", err)
			}
			return
		}

		if sh.Verbose {
			print(" -> %s", acct.Address)
		}

		err = acct.Update(sh, print)
		if err != nil {
			if !IsAccountDoesNotExist(err) {
				print("    updating from blockchain: %s", err.Error())
			}
			return
		}

		out <- acct
		// we know that by sending an account out the outbound channel, we're
		// about to trigger a new goroutine. We can't increment the WGs at the
		// site where we launch that goroutine, though, because there's a race
		// condition: if this is the last currently-existing goroutine, the
		// channel might close first. Instead, let's increment the WGs here:
		for _, wg := range wgs {
			wg.Add(1)
		}
	})
}
