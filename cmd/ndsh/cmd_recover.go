package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/alexflint/go-arg"

	"github.com/oneiro-ndev/ndaumath/pkg/words"

	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
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
		SeedWords   []string `arg:"positional,required" help:"seed phrase from which to recover this account"`
		Nicknames   []string `arg:"-n" help:"short nicknames which can refer to this account. Only applied if exactly one account was recovered"`
		Lang        string   `arg:"-l" help:"recovery phrase language"`
		Persistance int      `help:"number of non-accounts to discover before deciding there are no more in a derivation style"`
		Kind        string   `arg:"-k" help:"kind of account to attempt recovery of"`
	}{
		Lang:        "en",
		Persistance: 50,
		Kind:        string(address.KindUser),
	}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	if len(args.SeedWords) != 12 {
		fmt.Printf("WARN: ndau seed phrases are typically 12 words long, but you provided %d\n", len(args.SeedWords))
	}
	for idx := range args.SeedWords {
		(args.SeedWords)[idx] = strings.ToLower((args.SeedWords)[idx])
	}

	kind, err := address.ParseKind(args.Kind)
	if err != nil {
		return err
	}

	seed, err := words.ToBytes(args.Lang, args.SeedWords)
	if err != nil {
		return err
	}

	fmt.Println("Communicating with blockchain...")
	// TODO: add some kind of progress bar?
	// https://github.com/vbauerster/mpb seems good for this usage

	accountsStream := make(chan Account, 0)
	var wg sync.WaitGroup
	accounts := make([]Account, 0)

	// this should only ever be run in a new goroutine
	trypath := func(pattern string, idx int, ch chan<- Account) {
		path := fmt.Sprintf(pattern, idx)
		scopy := make([]byte, len(seed))
		copy(scopy, seed)
		buf := bytes.NewBuffer(scopy)
		wg.Add(1)
		sh.tryAccount(&wg, buf, path, kind, ch)
	}

	// for each account pattern, keep trying variations until persist failures
	for _, pattern := range accountPatterns {
		patstream := make(chan Account, 0)
		patmutex := sync.Mutex{}
		patidx := 0
		var patwg sync.WaitGroup
		// for every success, pass it to the outer stream, but also
		// try a new path as well
		go func() {
			for acct := range patstream {
				accountsStream <- acct
				patmutex.Lock()
				patidx++
				patmutex.Unlock()
				patwg.Add(1)
				go func() {
					trypath(pattern, patidx, patstream)
					patwg.Done()
				}()
			}
		}()

		// now generate the initial set of attempts for this pattern
		patmutex.Lock()
		for patidx = 0; patidx < args.Persistance; patidx++ {
			patwg.Add(1)
			go func(idx int) {
				trypath(pattern, idx, patstream)
				patwg.Done()
			}(patidx)
		}
		patmutex.Unlock()

		// close the pattern stream when we exhaust the pattern possibilities
		go func() {
			patwg.Wait()
			close(patstream)
		}()
	}

	// collect all results from all patterns into an array
	go func() {
		for acct := range accountsStream {
			accounts = append(accounts, acct)
		}
	}()

	wg.Wait()

	fmt.Printf("Discovered %d accounts:\n", len(accounts))
	for _, acct := range accounts {
		fmt.Printf("  %s (%s)\n", acct.Address, acct.Path)
		sh.accts.Add(acct)
	}
	// add nicknames if we've recovered exactly one account
	if len(accounts) == 1 && len(args.Nicknames) > 0 {
		fmt.Printf("Adding nicknames to %s: %s\n", accounts[0].Address, strings.Join(args.Nicknames, ", "))
		sh.accts.Add(accounts[0], args.Nicknames...)
	}

	return err
}

// try getting an account from the blockchain. If it exists, construct an appropriate
// struct and pass it along the channel. Discard any errors.
//
// Does _not_ attempt to discover any private keys
func (sh *Shell) tryAccount(
	wg *sync.WaitGroup,
	seed io.Reader,
	path string,
	kind byte,
	out chan<- Account,
) {
	defer wg.Done()

	acct, err := NewAccount(seed, path, kind)
	if err != nil {
		return
	}

	ad, resp, err := tool.GetAccount(sh.Node, acct.Address)
	if err != nil {
		return
	}
	exists := false
	_, err = fmt.Sscanf(resp.Response.Info, query.AccountInfoFmt, &exists)
	if err != nil || !exists {
		return
	}

	acct.Data = ad
}
