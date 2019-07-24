package main

// desired behaviors:

// test mode - everything to stderr, no prompt, output only on failure, simplified output, quit on err
// debug mode - prompts, input from stdin, all to stdout
// You're in one or the other.
// If you use the --verbose switch in test mode, it will drop to debug on err, output to stdout

// test mode is default if -s specified
// otherwise interactive mode

import (
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/chaincode/pkg/vm"
)

type args struct {
	Binary  string `arg:"-b" help:"File to load as a chasm binary (*.chbin)."`
	Script  string `arg:"-s" help:"Command script file (*.chasm) (sets test mode)."`
	Bytes   string `arg:"-B" help:"Raw chaincode bytes to preload as a script. Must be space-separated base10 unless --hex-bytes or --base64-bytes is set."`
	Base64  bool   `arg:"--base64-bytes" help:"Interpret --bytes input as base64-encoded"`
	Hex     bool   `arg:"--hex-bytes" help:"Interpret --bytes input as hex-encoded"`
	Verbose bool   `arg:"-v" help:"Verbose output; errors in test mode will drop into debug mode."`
	Test    bool   `arg:"-t" help:"Forces test mode."`
	Debug   bool   `arg:"-d" help:"Forces debug mode."`
}

func (args) Description() string {
	return `crank is a repl for chaincode.

	crank starts up and creates a new VM with no contents.

	If --binary was specified, crank attempts to load the given file instead of starting with an empty vm.
	if --script was specified, crank then attempts to read the specified file as if it were a series of
	commands, and sets test mode.

	If --bytes was specified, crank interprets the argument as a compiled chaincode script. If --base64-bytes
	was also specified, crank interprets the argument as base64-encoded data. If --hex-bytes was also
	specified, crank interprets the argument as hexadecimal data, ignoring spaces. Otherwise, crank
	interprets the input as space-separated base10 values.

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

	switch {
	case a.Script != "":
		inf, err := os.Open(a.Script)
		if err != nil {
			log.Fatalf("Unable to load script file %s: %s", a.Script, err)
		}
		rs.script = a.Script
		rs.in = inf
		rs.mode = TEST
	case a.Binary != "":
		err := rs.load(a.Binary)
		if err != nil {
			log.Fatalf("Unable to load binary file %s: %s", a.Binary, err)
		}
	case a.Bytes != "":
		// strip enclosing brackets if present
		if a.Bytes[0] == '[' && a.Bytes[len(a.Bytes)-1] == ']' {
			a.Bytes = a.Bytes[1 : len(a.Bytes)-1]
		}

		// decode bytes
		var bytes []byte
		var err error
		switch {
		case a.Base64:
			bytes, err = base64.StdEncoding.DecodeString(a.Bytes)
		case a.Hex:
			// strip spaces from hex input so it all parses properly
			a.Bytes = strings.Replace(a.Bytes, " ", "", -1)
			bytes, err = hex.DecodeString(a.Bytes)
		default:
			numbers := strings.Fields(a.Bytes)
			bytes = make([]byte, 0, len(numbers))
			for _, numberS := range numbers {
				var number uint64
				// can't use the handy '0' base here because it would interpret
				// leading 0 characters as octal, which might lead to surprising
				// behavior
				number, err = strconv.ParseUint(numberS, 10, 8)
				if err != nil {
					break
				}
				bytes = append(bytes, byte(number))
			}
		}
		if err != nil {
			log.Fatalf("Unable to parse bytes input: %s", err)
		}

		var cvm *vm.ChaincodeVM
		cvm, err = vm.NewChaincode(vm.ToChaincode(bytes))
		if err != nil {
			log.Fatalf("Unable to construct raw vm: %s", err)
		}
		rs.vm = cvm.MakeMutable()

		if a.Verbose {
			rs.dispatch("dis")
		}

		rs.vm.Init(0)
	}

	switch {
	case a.Debug:
		rs.mode = DEBUG
	case a.Test:
		rs.mode = TEST
	}

	rs.verbose = a.Verbose

	rs.repl()
}
