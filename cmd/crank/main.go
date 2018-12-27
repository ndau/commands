package main

// desired behaviors:

// test mode - everything to stderr, no prompt, output only on failure, simplified output, quit on err
// debug mode - prompts, input from stdin, all to stdout
// You're in one or the other.
// If you use the --verbose switch in test mode, it will drop to debug on err, output to stdout

// test mode is default if -i specified
// otherwise interactive mode

import (
	"log"
	"os"

	arg "github.com/alexflint/go-arg"
)

type args struct {
	Binary  string `arg:"-b" help:"File to load as a chasm binary (*.chbin)."`
	Script  string `arg:"-s" help:"Command script file (sets test mode)."`
	Verbose bool   `arg:"-v" help:"Verbose output; errors in test mode will drop into debug mode."`
	Test    bool   `arg:"-t" help:"Forces test mode."`
	Debug   bool   `arg:"-d" help:"Forces debug mode."`
}

func (args) Description() string {
	return `crank is a repl for chaincode.

	crank starts up and creates a new VM with no contents.

	If --binary was specified, crank attempts to load the given file instead of starting with an empty vm.
	if --script was specified, crank then attempts to read the specified file as if it were a series of
	commands. If you want crank to terminate automatically, make sure your input file ends in a quit command.

	crank has two modes: debug and test. In test mode, it expects to load a script file (and loading
	one on the command line automatically sets test mode). Errors go to stdout and only errors are emitted.
	A clean test exits with no output and errorlevel 0. A failed test has output and errorlevel >0.

	In debug mode, it's an interactive repl.

	You can also set a verbose flag, which prints lots of stuff. In test mode, an error in verbose mode
	causes crank to drop into the console.
	`
}

func main() {
	// this needs to be filled in dynamically because the help function traverses
	// the commands list.
	h := commands["help"]
	h.handler = help
	commands["help"] = h

	var a args

	arg.MustParse(&a)
	rs := runtimeState{mode: DEBUG, in: os.Stdin, out: newOutputter()}
	if a.Script != "" {
		inf, err := os.Open(a.Script)
		if err != nil {
			log.Fatalf("Unable to load script file %s: %s", a.Script, err)
		}
		rs.script = a.Script
		rs.in = inf
		rs.mode = TEST
	}
	if a.Debug {
		rs.mode = DEBUG
	}
	if a.Test {
		rs.mode = TEST
	}
	rs.verbose = a.Verbose

	if a.Binary != "" {
		err := rs.load(a.Binary)
		if err != nil {
			log.Fatalf("Unable to load binary file %s: %s", a.Binary, err)
		}
	}

	rs.repl()
}
