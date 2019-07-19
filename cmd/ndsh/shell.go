package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/shlex"
	"github.com/mitchellh/go-homedir"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
)

// Shell manages global state, dispatching commands, and other similar responsibilities.
type Shell struct {
	Commands map[string]Command
	Ps1      string
	Running  sync.WaitGroup
	Stop     chan struct{}
	Node     client.ABCIClient
	Verbose  bool
	Staged   *Stage
	Accts    *Accounts

	ireader     *bufio.Reader
	writelock   sync.Mutex
	writer      *bufio.Writer
	systemAccts map[string]string
}

// NewShell initializes the shell
func NewShell(verbose bool, node client.ABCIClient, commands ...Command) *Shell {
	sh := Shell{
		Commands: make(map[string]Command),
		Ps1:      "{tx}ndsh> ",
		Stop:     make(chan struct{}),
		Node:     node,
		Verbose:  verbose,
		ireader:  bufio.NewReader(os.Stdin),
		Accts:    NewAccounts(),
		writer:   bufio.NewWriter(os.Stdout),
	}
	for _, command := range commands {
		for _, name := range strings.Split(command.Name(), " ") {
			sh.Commands[name] = command
		}
	}
	if ps1 := os.ExpandEnv("$NDSH_PS1"); len(ps1) > 0 {
		sh.Ps1 = ps1
	}
	return &sh
}

func timeout(f func(), d time.Duration, m string) {
	ch := make(chan struct{})
	go func() {
		f()
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		// cool, everything is fine
	case <-time.After(d):
		panic(m)
	}
}

// Exit the shell, shutting down all async commands
//
// If err is not nil, write it to stderr and set a non-0 exit code.
// Otherwise, write nothing and return 0.
func (sh *Shell) Exit(err error) {
	close(sh.Stop)
	// wait for all the goroutines to close gracefully, or panic
	timeout(func() { sh.Running.Wait() }, 5*time.Second, "misbehaved goroutines didn't stop!")

	// if all subcommands have shut down, we can exit gracefully
	check(err, "error")
	os.Exit(0)
}

// this is just a stub for now, but the intent is to be able to expand variables
// into the ndau shell's prompt
func (sh *Shell) expandPrompt() string {
	prompt := sh.Ps1
	for _, subs := range []struct {
		replace string
		with    func() string
	}{
		{"{tx}", sh.staged},
	} {
		prompt = strings.Replace(prompt, subs.replace, subs.with(), -1)
	}
	return prompt
}

func (sh *Shell) staged() string {
	if sh.Staged == nil || sh.Staged.Tx == nil {
		return ""
	}
	return fmt.Sprintf("(%s) ", metatx.NameOf(sh.Staged.Tx))
}

// prompt the user, and dispatch appropriate commands
//
// For now, we're using cooked line disciplines, meaning that we can't
// yet intercept arrows (for in-memory history) or tabs (for completion).
// This is what we'll want to mess with to change that in the future.
func (sh *Shell) prompt() {
	// we can't use sh.Write here, because we don't want a newline.
	// However, we still want to ensure that we wait until the lock is ready.
	sh.writelock.Lock()
	fmt.Print(sh.expandPrompt())
	sh.writelock.Unlock()
	input, err := sh.ireader.ReadString('\n')
	check(err, "scanning input line from user")
	err = sh.Exec(input)
	if err != nil {
		sh.Write(err.Error())
	}
}

// Exec runs the command per a given input
func (sh *Shell) Exec(input string) error {
	var err error
	commands := strings.Split(input, "&&")
	for _, command := range commands {
		var tokens []string
		tokens, err = shlex.Split(command)
		check(err, "tokenizing user input")
		if len(tokens) > 0 {
			cmd := sh.Commands[tokens[0]]
			if cmd == nil {
				return fmt.Errorf("command not found: '%s'", tokens[0])
			}
			err = cmd.Run(tokens, sh)
		}
		if err != nil {
			break
		}
	}
	return err
}

// Run the shell
func (sh *Shell) Run() {
	for {
		sh.prompt()
	}
}

// Write some data to the shell's output
func (sh *Shell) Write(format string, context ...interface{}) {
	if format[len(format)-1] != '\n' {
		format += "\n"
	}
	sh.writelock.Lock()
	defer sh.writelock.Unlock()
	fmt.Fprintf(sh.writer, format, context...)
	sh.writer.Flush()
}

// WriteBatch writes connected messages to the shell's output, ensuring it's not interrupted by other routines
func (sh *Shell) WriteBatch(writes func(print func(format string, context ...interface{}))) {
	sh.writelock.Lock()
	defer sh.writelock.Unlock()
	writes(func(format string, context ...interface{}) {
		if format == "" || format[len(format)-1] != '\n' {
			format += "\n"
		}
		fmt.Fprintf(sh.writer, format, context...)
	})
	sh.writer.Flush()
}

// VWrite writes the message if the shell is in Verbose mode
func (sh *Shell) VWrite(format string, context ...interface{}) {
	if sh.Verbose {
		sh.Write(format, context...)
	}
}

// LoadSystemAccts from system_accts.toml
func (sh *Shell) LoadSystemAccts(path string) (err error) {
	path = os.ExpandEnv(path)
	path, err = homedir.Expand(path)
	if err != nil {
		return errors.Wrap(err, "expanding homedir")
	}

	_, err = toml.DecodeFile(path, &sh.systemAccts)
	if err != nil {
		return errors.Wrap(err, "decoding "+path)
	}
	return nil
}

func (sh *Shell) saStr(sv string) (string, error) {
	if sh.systemAccts == nil {
		return "", errors.New("system accounts not loaded")
	}
	s, ok := sh.systemAccts[sv]
	if !ok {
		return "", errors.New(sv + " not found in system accounts")
	}
	return s, nil
}

// SAAddr returns the system account address for a given system variable
func (sh *Shell) SAAddr(sv string) (*address.Address, error) {
	s, err := sh.saStr(sv)
	if err != nil {
		return nil, err
	}
	a, err := address.Validate(s)
	return &a, err
}

// SAPrivateKey returns the private key for a given system variable
func (sh *Shell) SAPrivateKey(sv string) (*signature.PrivateKey, error) {
	s, err := sh.saStr(sv)
	if err != nil {
		return nil, err
	}
	return signature.ParsePrivateKey(s)
}

// SAPublicKey returns the public key for a given system variable
func (sh *Shell) SAPublicKey(sv string) (*signature.PublicKey, error) {
	s, err := sh.saStr(sv)
	if err != nil {
		return nil, err
	}
	return signature.ParsePublicKey(s)
}

// SAAcct returns an Account for the magic account for a system variable.
//
// Does not add this account to the accounts list
func (sh *Shell) SAAcct(addressName string, validationPrivateName string) (*Account, error) {
	addr, err := sh.SAAddr(addressName)
	if err != nil {
		return nil, errors.Wrap(err, "getting address for magic account")
	}

	pvt, err := sh.SAPrivateKey(validationPrivateName)
	if err != nil {
		return nil, err
	}

	acct := &Account{
		Address: *addr,
		PrivateValidationKeys: []signature.PrivateKey{
			*pvt,
		},
	}
	err = acct.Update(sh, sh.Write)
	err = errors.Wrap(err, "updating magic account")
	return acct, err
}

// Dispatch a tx, handling staging appropriately
func (sh *Shell) Dispatch(stage bool, tx metatx.Transactable, update, magic *Account) error {
	sh.Staged = &Stage{
		Tx: tx,
	}
	if magic != nil {
		sh.Staged.Account = magic
	} else if update != nil {
		sh.Staged.Account = update
	}

	if stage {
		return nil
	}

	if s, ok := tx.(ndau.Signeder); ok {
		if len(s.GetSignatures()) == 0 {
			return errors.New("tx has 0 signatures; will not send")
		}
	}

	sh.VWrite("sending tx with hash %s", metatx.Hash(tx))

	_, err := tool.SendCommit(sh.Node, tx)
	if err != nil {
		return errors.Wrap(err, "sending transaction")
	}

	// clear staging after successful send
	sh.Staged = nil

	if update != nil {
		err = update.Update(sh, sh.Write)
	}
	return err
}

// AddressOf returns the address from an input string
//
// This extends sh.Accts.Get: it has all the same behavior,
// plus the additional benefit that if the input isn't a known
// account but it is a complete address, it succeeds.
//
// Exactly one of the address and error return values will always be nil.
// If the account is known, it will be returned.
func (sh *Shell) AddressOf(s string) (*address.Address, *Account, error) {
	acct, err := sh.Accts.Get(s)
	switch {
	case err == nil:
		return &acct.Address, acct, nil
	case IsNoMatch(err):
		addr, err := address.Validate(s)
		if err != nil {
			return nil, nil, errors.Wrap(err, "input must be a complete address or a suffix of a known account address or nickname")
		}
		return &addr, nil, nil
	default:
		return nil, nil, err
	}
}
