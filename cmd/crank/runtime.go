package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
)

// Mode could theoretically be a bool but we want to reserve the option to have
// more states.
type Mode int

// These are constants for Mode
const (
	DEBUG Mode = iota
	TEST  Mode = iota
)

type runtimeState struct {
	vm      *vm.ChaincodeVM
	event   byte
	stack   *vm.Stack
	binary  string
	script  string
	mode    Mode
	verbose bool
	lastcmd string
	in      io.Reader
	out     *outputter
}

func help(rs *runtimeState, args string) error {
	keys := make([]string, 0, len(commands))
	for key := range commands {
		keys = append(keys, key)
	}
	sort.Sort(sort.StringSlice(keys))
	for _, key := range keys {
		rs.out.Printf("%11s: %s %s\n", key, commands[key].summary, commands[key].aliases)
		if strings.HasPrefix(args, "v") {
			rs.out.Println(commands[key].detail)
		}
	}
	return nil
}

// load is a command that loads a file into a VM (or errors trying)
func (rs *runtimeState) load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		// if we failed to open, it might be because the binary is relative to the script
		if filepath.IsAbs(filename) || rs.script == "" {
			return newExitError(1, err, nil)
		}
		// try to see if we can assemble a path relative to the script dir
		scriptdir := filepath.Dir(rs.script)
		fp := filepath.Join(scriptdir, filename)
		f, err = os.Open(fp)
		if err != nil {
			return newExitError(1, err, nil)
		}
	}
	bin, err := vm.Deserialize(f)
	if err != nil {
		return newExitError(1, err, nil)
	}
	vm, err := vm.New(bin)
	rs.vm = vm
	rs.binary = filename
	if err != nil {
		return newExitError(1, err, rs)
	}
	vm.Init(0)
	return nil
}

// reinit calls init again, duplicating the entries that are currently
// on the stack.
func (rs *runtimeState) reinit(stk *vm.Stack) error {
	// copy the current stack and save it in case we need to reset
	rs.stack = stk.Clone()

	// now initialize
	return rs.vm.InitFromStack(rs.event, rs.stack)
}

// setevent sets up the VM to run the given event, which means that it calls
// reinit to set up the stack as well.
func (rs *runtimeState) setevent(eventid string) error {
	if v, ok := predefined[eventid]; ok {
		eventid = v
	}
	ev, err := strconv.ParseInt(strings.TrimSpace(eventid), 0, 8)
	if err != nil {
		return err
	}
	rs.event = byte(ev)

	return rs.reinit(rs.vm.Stack())
}

func (rs *runtimeState) run(debug vm.Dumper) error {
	err := rs.vm.Run(debug)
	return err
}

func (rs *runtimeState) step(debug vm.Dumper) error {
	err := rs.vm.Step(debug)
	return err
}

var p = regexp.MustCompile("[[:space:]]+")

func (rs *runtimeState) dispatch(s string) error {
	if rs.vm == nil {
		var err error
		rs.vm, err = vm.NewEmpty()
		if err != nil {
			return err
		}
		rs.vm.Init(0)
	}
	args := p.Split(s, 2)
	for key, cmd := range commands {
		if key == args[0] || cmd.matchesAlias(args[0]) {
			extra := ""
			if len(args) > 1 {
				extra = args[1]
			}
			return cmd.handler(rs, extra)
		}
	}
	return fmt.Errorf("unknown command %s - type ? for help", s)
}

func stripComments(s string) string {
	s = strings.TrimSpace(s)
	ix := strings.Index(s, ";")
	if ix != -1 {
		return s[:ix]
	}
	return s
}

func (rs *runtimeState) repl() {
	reader := bufio.NewReader(rs.in)
	for linenumber := 1; ; linenumber++ {
		// prompt always preceded by current vm data
		if rs.verbose || rs.mode == DEBUG {
			if rs.vm == nil {
				rs.out.Println("  [no VM is loaded]")
			} else {
				rs.out.Println(rs.vm)
			}
			rs.out.Printf("%3d crank> ", linenumber)
		}
		// get one line
		if rs.verbose || rs.mode == DEBUG {
			rs.out.Flush(os.Stdout, true)
		}
		inputline, err := reader.ReadString('\n')
		// if an empty line was received and we're in debug mode, it means repeat previous line
		if rs.mode == DEBUG && inputline == "\n" {
			inputline = rs.lastcmd + "\n"
		}

		// echo if that line came from an input file as it was not echoed on input
		if rs.verbose && rs.mode == TEST {
			rs.out.Printf("%s", inputline)
		}
		// look for errors on input
		if err != nil && err != io.EOF {
			rs.out.Errorln(err)
			rs.out.Flush(os.Stderr, rs.mode == DEBUG)
			os.Exit(1)
		}
		// eof from terminal means quit
		if err == io.EOF && rs.mode == DEBUG {
			// we're really done now, shut down normally by injecting a quit command
			inputline = "quit\n"
			err = nil
		}
		// eof from input might mean exit, or it might mean drop into stdin
		// if we're in verbose mode
		if err == io.EOF && rs.mode == TEST {
			if rs.verbose {
				reader = bufio.NewReader(os.Stdin)
				rs.mode = DEBUG
				rs.out.Println("*** Input now from stdin ***")
			} else {
				// nope, we're done
				os.Exit(0)
			}
		}
		// ignore blank lines and comments
		inputline = stripComments(inputline)
		if inputline == "" {
			continue
		}

		// we got an attempt at a command, record it
		rs.lastcmd = inputline

		// now try it
		err = rs.dispatch(inputline)
		switch e := err.(type) {
		case exiter:
			if e.Error() != "" {
				rs.out.Errorf("%s: line %d: error: %s\n", rs.script, linenumber, e.Error())
			} else {
				rs.out.Printf("%s: %d lines.\n", rs.script, linenumber)
			}
			// fmt.Println("inside exiter case", rs.verbose || rs.mode == DEBUG, string(rs.out.rows[0].content))
			rs.out.Flush(os.Stdout, rs.verbose || rs.mode == DEBUG)
			if rs.verbose && rs.mode == TEST {
				reader = bufio.NewReader(os.Stdin)
				rs.mode = DEBUG
				rs.out.Println("*** Exit requested while verbose: input now from stdin ***")
			} else {
				e.Exit()
			}
		case error:
			rs.out.Errorf("line %d: error: %s\n", linenumber, err)
		case nil:
		default:
		}
		if rs.vm != nil && (rs.verbose && rs.mode == DEBUG) {
			rs.vm.Disassemble(rs.vm.IP())
		}
	}
}
