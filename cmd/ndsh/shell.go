package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/shlex"
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

	ireader   *bufio.Reader
	accts     *Accounts
	writelock sync.Mutex
	writer    *bufio.Writer
}

// NewShell initializes the shell
func NewShell(verbose bool, node client.ABCIClient, commands ...Command) *Shell {
	sh := Shell{
		Commands: make(map[string]Command),
		Ps1:      "ndsh> ",
		Stop:     make(chan struct{}),
		Node:     node,
		Verbose:  verbose,
		ireader:  bufio.NewReader(os.Stdin),
		accts:    NewAccounts(),
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
	return sh.Ps1
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
		if format[len(format)-1] != '\n' {
			format += "\n"
		}
		fmt.Fprintf(sh.writer, format, context...)
	})
	sh.writer.Flush()
}
