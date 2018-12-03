package main

import (
	"errors"
	"fmt"
	"os"
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
		handler: nil, //  we need to fill this in dynamically because it traverses this list
	},
	"quit": command{
		aliases: []string{"q"},
		summary: "ends the chain program",
		detail:  `Ctrl-D also works`,
		handler: func(rs *runtimeState, args string) error {
			fmt.Println("Goodbye.")
			os.Exit(0)
			return nil
		},
	},
	"exit": command{
		aliases: []string{"q"},
		summary: "pops the stack; if the top of stack was numeric, uses the lowest byte of its value as the OS exit level",
		detail:  `If the top of stack did not exist or was not numeric, exits with 255.`,
		handler: func(rs *runtimeState, args string) error {
			n, err := rs.vm.Stack().PopAsInt64()
			if err != nil {
				os.Exit(255)
			}
			os.Exit(int(n & 0xFF))
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
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			return rs.run(false)
		},
	},
	"next": command{
		aliases: []string{"n"},
		summary: "executes one opcode at the current IP and prints the status",
		detail:  `If the opcode is a function call, this executes the entire function call before stopping.`,
		handler: func(rs *runtimeState, args string) error {
			return rs.step(true)
		},
	},
	"trace": command{
		aliases: []string{"tr", "t"},
		summary: "runs the currently loaded VM from the current IP",
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			return rs.run(true)
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
			rs.vm.DisassembleAll()
			return nil
		},
	},
	"reset": command{
		aliases: []string{"k"},
		summary: "resets the VM to the event and stack that were current at the last Run, Trace, Push, Pop, or Event command",
		detail:  ``,
		handler: func(rs *runtimeState, args string) error {
			rs.reinit(rs.stack)
			fmt.Println(rs.vm.Stack())
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
    Quoted string
    B(hex pairs)
    [ list of values ] (commas or whitespace, must all be one line)
    { list of values }(commas or whitespace, must all be one line)
		`,
		handler: func(rs *runtimeState, args string) error {
			topush, err := parseValues(args)
			if err != nil {
				return err
			}
			for _, v := range topush {
				rs.vm.Stack().Push(v)
			}
			fmt.Println(rs.vm.Stack())
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
}
