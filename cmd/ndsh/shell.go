package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/shlex"
)

// Shell manages global state, dispatching commands, and other similar responsibilities.
type Shell struct {
	Commands map[string]Command
	Ps1      string
	Running  sync.WaitGroup
	Stop     chan struct{}

	ireader *bufio.Reader
}

// NewShell initializes the shell
func NewShell(commands ...Command) *Shell {
	sh := Shell{
		Commands: make(map[string]Command),
		Ps1:      "ndsh> ",
		Stop:     make(chan struct{}),
		ireader:  bufio.NewReader(os.Stdin),
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
	check(err)
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
	fmt.Print(sh.expandPrompt())
	input, err := sh.ireader.ReadString('\n')
	checkw(err, "scanning input line from user")
	tokens, err := shlex.Split(input)
	checkw(err, "tokenizing user input")
	if len(tokens) > 0 {
		cmd := sh.Commands[tokens[0]]
		if cmd == nil {
			fmt.Printf("command not found: '%s'\n", tokens[0])
			return
		}
		cmd.Run(tokens, sh)
	}
}

// Run the shell
func (sh *Shell) Run() {
	for {
		sh.prompt()
	}
}
