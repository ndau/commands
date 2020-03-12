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
	"os"
	"strconv"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/ndau/chaincode/pkg/vm"
	"github.com/pkg/errors"
)

// command is a type that is used to create a table of commands for the repl
// yes, we could do this by mapping all the names and aliases to a single map,
// but it's useful for help to have a difference between the names and the aliases
type command struct {
	parms   string
	aliases []string
	summary string
	detail  string
	handler func(rs *runtimeState, args string) error
}

func (c command) matchesAlias(s string) bool {
	for _, a := range c.aliases {
		if s == a {
			return true
		}
	}
	return false
}

var commands = map[string]command{
	"help": command{
		aliases: []string{"?"},
		summary: "prints this help message (help verbose for extended explanation)",
		detail:  ``,
		handler: nil, //  we need to fill this in dynamically because the handler
		// traverses this list; a static assignment causes a reference loop
	},
	"quit": command{
		aliases: []string{"q"},
		summary: "ends the chain program",
		detail:  `Ctrl-D also works`,
		handler: func(rs *runtimeState, args string) error {
			return newExitError(0, nil, nil)
		},
	},
	"exit": command{
		aliases: []string{},
		summary: "pops the stack; if the top of stack was numeric, uses the lowest byte of its value as the OS exit level",
		detail:  `If the top of stack did not exist or was not numeric, exits with 255.`,
		handler: func(rs *runtimeState, args string) error {
			n, err := rs.vm.Stack().PopAsInt64()
			if err != nil {
				return newExitError(255, err, nil)
			}
			return newExitError(int(n&0xFF), nil, rs)
		},
	},
	"expect": command{
		aliases: []string{},
		summary: "Compares stack to the given value(s).",
		detail:  expectParser{}.Description(),
		handler: func(rs *runtimeState, args string) error {
			ep := expectParser{}
			values, err := ep.Parse(args, rs)
			if err != nil {
				return newExitError(255, err, rs)
			}
			compare, err := ep.Comparitor(rs)
			if err != nil {
				return newExitError(255, err, rs)
			}
			for _, v := range values {
				stk, err := rs.vm.Stack().Pop()
				if err != nil {
					return newExitError(255, err, rs)
				}
				if err = compare(stk, v); err != nil {
					return newExitError(1, err, rs)
				}
			}
			return nil
		},
	},
	"load": command{
		aliases: []string{"l"},
		summary: "loads the file FILE as a chasm binary (.chbin)",
		detail:  `File must conform to the chasm binary standard.`,
		handler: (*runtimeState).load,
	},
	"run": command{
		aliases: []string{"r"},
		summary: "runs the currently loaded VM from the current IP",
		detail:  `if arg is "fail" or "succeed" will exit if the result disagrees`,
		handler: func(rs *runtimeState, rargs string) error {
			var dumper vm.Dumper
			if args.Verbose {
				dumper = vm.Trace
			}
			err := rs.run(dumper)
			switch strings.ToLower(rargs) {
			case "fail":
				if err == nil {
					val, err := rs.vm.Stack().PopAsInt64()
					if err == nil && val == 0 {
						return newExitError(1, errors.New("expected to fail, but didn't"), rs)
					}
				}
				return nil // we expected to fail, so we're happy about that
			case "succeed", "success":
				if err != nil {
					return newExitError(2, fmt.Errorf("expected to succeed, but failed (%s)", err), rs)
				}
				val, err := rs.vm.Stack().PopAsInt64()
				if err != nil {
					return newExitError(2, fmt.Errorf("expected to succeed, but failed (%s)", err), rs)
				}
				if val != 0 {
					return newExitError(3, fmt.Errorf("expected to succeed, but returned %d", val), rs)
				}
			}
			return err
		},
	},
	"next": command{
		aliases: []string{"n"},
		summary: "executes one opcode at the current IP and prints the status",
		detail:  `If the opcode is a function call, this executes the entire function call before stopping.`,
		handler: func(rs *runtimeState, args string) error {
			dumper := func(vm *vm.ChaincodeVM) {
				rs.out.Println(vm)
			}
			return rs.step(dumper)
		},
	},
	"trace": command{
		aliases: []string{"tr", "t"},
		summary: "runs the currently loaded VM from the current IP",
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			dumper := func(vm *vm.ChaincodeVM) {
				rs.out.Println(vm)
			}
			return rs.run(dumper)
		},
	},
	"event": command{
		aliases: []string{"ev", "e"},
		summary: "sets the ID of the event to be executed (may change the current IP)",
		detail:  ``,
		handler: (*runtimeState).setevent,
	},
	"disassemble": command{
		aliases: []string{"dis", "disasm", "d"},
		summary: "disassembles the loaded vm",
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			if rs.vm == nil {
				return errors.New("no VM is loaded")
			}
			rs.vm.DisassembleAll(os.Stdout)
			return nil
		},
	},
	"reset": command{
		aliases: []string{},
		summary: "resets the VM to the event and stack that were current at the last Run, Trace, Push, Pop, or Event command",
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			rs.reinit(rs.stack)
			fmt.Println(rs.vm.Stack())
			return nil
		},
	},
	"clear": command{
		aliases: []string{},
		summary: "clears the stack",
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			rs.reinit(vm.NewStack())
			return nil
		},
	},
	"stack": command{
		aliases: []string{"k"},
		summary: "prints the contents of the stack",
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			fmt.Println(rs.vm.Stack())
			return nil
		},
	},
	"push": command{
		aliases: []string{"pu", "p"},
		summary: "pushes one or more values onto the stack",
		detail: `
Value syntax:
    Number (decimal, hex)
    Timestamp
    Quoted string (converted to bytes)
    B(hex pairs)
    [ list of values ] (commas or whitespace, must all be one line)
    { struct         } (commas or whitespace, must all be one line)
		`,
		handler: func(rs *runtimeState, args string) error {
			topush, err := parseValues(args)
			if err != nil {
				return err
			}
			for _, v := range topush {
				rs.vm.Stack().Push(v)
			}
			return rs.reinit(rs.vm.Stack())
		},
	},
	"pop": command{
		aliases: []string{"o"},
		summary: "pops the top stack item and prints it",
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			v, err := rs.vm.Stack().Pop()
			if err != nil {
				return err
			}
			fmt.Println(v)
			return rs.reinit(rs.vm.Stack())
		},
	},
	"constants": command{
		aliases: []string{"const"},
		summary: "prints the list of predefined constants (restricting to those containing a substring if specified)",
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			for k := range predefined {
				if args == "" || strings.Contains(k, strings.ToUpper(args)) {
					fmt.Println(k)
				}
			}
			return nil
		},
	},
	"set-now": command{
		aliases: []string{"setnow", "sn"},
		summary: "sets the value which the vm will return for the `now` opcode",
		detail: `
argument must be RFC3339 format timestamp, or empty

empty argument means that the VM should return the current time
`,
		handler: func(rs *runtimeState, args string) error {
			args = strings.TrimSpace(args)
			var n vm.Nower
			if args == "" {
				var err error
				n, err = vm.NewDefaultNow()
				if err != nil {
					return errors.Wrap(err, "constructing default now")
				}
			} else {
				ts, err := vm.ParseTimestamp(args)
				if err != nil {
					return errors.Wrap(err, "parsing arg as timestamp")
				}
				n = nower{ts}
			}
			rs.vm.SetNow(n)
			return nil
		},
	},
}

type expectParser struct {
	Delta   string   `arg:"-d,--delta" help:"absolute allowed error"`
	Epsilon string   `arg:"-e,--epsilon" help:"relative allowed error"`
	Values  []string `arg:"positional"`
}

func (expectParser) Description() string {
	return `
If the expected values are not found or an error occurs, exits with a nonzero
return code.

If --delta is set, the DELTA value is the absolute allowed error. For example:

	expect 100 --delta 1

In this case, the expect will succeed if the stack head is in the
inclusive range [99..101].

If --epsilon is set, the EPSILON value is the relative allowed error. For example:

	expect 100 --epsilon 0.1

In this case, the expect will succeed if the stack head is in the
inclusive range [90..110].

--delta and --epsilon are mutually exclusive. Both options apply to all values
in the stack.
	`
}

func (ep expectParser) vmValues() ([]vm.Value, error) {
	out := make([]vm.Value, 0, len(ep.Values))
	for _, v := range ep.Values {
		vs, err := parseValues(v)
		if err != nil {
			return out, err
		}
		out = append(out, vs...)
	}
	return out, nil
}

// Parse consumes an argument string and returns a list of values
func (ep *expectParser) Parse(args string, rs *runtimeState) ([]vm.Value, error) {
	parser, err := arg.NewParser(arg.Config{}, ep)
	if err != nil {
		return nil, newExitError(254, err, rs)
	}
	err = parser.Parse(strings.Split(args, " "))
	if err != nil {
		return nil, newExitError(255, err, rs)
	}
	if ep.Delta != "" && ep.Epsilon != "" {
		return nil, newExitError(255, errors.New("--delta and --epsilon are incompatible"), rs)
	}
	return ep.vmValues()
}

func (ep expectParser) Comparitor(rs *runtimeState) (func(have, want vm.Value) error, error) {
	if ep.Delta != "" {
		delta, err := strconv.ParseInt(ep.Delta, 0, 64)
		if err != nil {
			return nil, newExitError(255, err, rs)
		}
		return func(have, want vm.Value) error {
			hn, ok := have.(vm.Numeric)
			if !ok {
				return fmt.Errorf("%s (on stack) is not numeric", have)
			}
			wn, ok := want.(vm.Numeric)
			if !ok {
				return fmt.Errorf("%s (given) is not numeric", want)
			}

			d := hn.AsInt64() - wn.AsInt64()
			if d < 0 {
				d = -d
			}
			if d > delta {
				return fmt.Errorf("%s (on stack) not within delta %v of %s (given) - exiting", have, delta, want)
			}
			return nil
		}, nil
	}

	if ep.Epsilon != "" {
		epsilon, err := strconv.ParseFloat(ep.Epsilon, 64)
		if err != nil {
			return nil, newExitError(255, err, rs)
		}
		return func(have, want vm.Value) error {
			hn, ok := have.(vm.Numeric)
			if !ok {
				return fmt.Errorf("%s (on stack) is not numeric", have)
			}
			wn, ok := want.(vm.Numeric)
			if !ok {
				return fmt.Errorf("%s (given) is not numeric", want)
			}

			h := float64(hn.AsInt64())
			w := float64(wn.AsInt64())
			d := h - w
			if d < 0 {
				d = -d
			}
			e := w * epsilon
			if d > e {
				return fmt.Errorf("%s (on stack) not within epsilon %v of %s (given) - exiting", have, epsilon, want)
			}
			return nil
		}, nil
	}

	return func(have, want vm.Value) error {
		if !want.Equal(have) {
			return fmt.Errorf("%s (on stack) does not equal %s (given) - exiting", have, want)
		}
		return nil
	}, nil
}
