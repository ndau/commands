package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/oneiro-ndev/chaincode/pkg/vm"

	arg "github.com/alexflint/go-arg"
)

// crank is a repl for chaincode

// crank -i inputstream -f FILE.chbin

// crank starts up and creates a new VM with no contents
// If -f was specified, crank attempts to load the given file instead of starting with an empty vm
// if -i was specified, crank then attempts to read the input file as if it were a series of
// commands. When -i terminates, it returns control to the normal input. If you want crank to
// terminate automatically, make sure your input file ends in a quit command.

type runtimeState struct {
	vm    *vm.ChaincodeVM
	event byte
	stack *vm.Stack
}

func help(rs *runtimeState, args string) error {
	keys := make([]string, 0, len(commands))
	for key := range commands {
		keys = append(keys, key)
	}
	sort.Sort(sort.StringSlice(keys))
	for _, key := range keys {
		fmt.Printf("%11s: %s %s\n", key, commands[key].summary, commands[key].aliases)
		if strings.HasPrefix(args, "v") {
			fmt.Println(commands[key].detail)
		}
	}
	return nil
}

// load is a command that loads a file into a VM (or errors trying)
func (rs *runtimeState) load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	bin, err := vm.Deserialize(f)
	if err != nil {
		return err
	}
	vm, err := vm.New(bin)
	if err != nil {
		return err
	}
	vm.Init(0)
	rs.vm = vm
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

func (rs *runtimeState) run(debug bool) error {
	err := rs.vm.Run(debug)
	return err
}

func (rs *runtimeState) step(debug bool) error {
	err := rs.vm.Step(debug)
	return err
}

func (rs *runtimeState) dispatch(s string) error {
	p := regexp.MustCompile("[[:space:]]+")
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

func (rs *runtimeState) repl(cmdsrc io.Reader, verbose bool) {
	reader := bufio.NewReader(os.Stdin)
	usingStdin := true
	if cmdsrc != nil {
		reader = bufio.NewReader(cmdsrc)
		usingStdin = false
	}

	for linenumber := 1; ; linenumber++ {
		// prompt always preceded by current vm data
		if verbose || usingStdin {
			if rs.vm == nil {
				fmt.Println("  [no VM is loaded]")
			} else {
				fmt.Println(rs.vm)
			}
			fmt.Printf("%3d crank> ", linenumber)
		}
		// get one line
		s, err := reader.ReadString('\n')
		// if that line came from outside, echo it
		if !usingStdin && verbose {
			fmt.Print(s)
		}
		// look for errors
		if err != nil && err != io.EOF {
			panic(err)
		}
		// eof from terminal means quit
		if err == io.EOF && usingStdin == true {
			// we're really done now, shut down normally
			s = "quit\n"
			err = nil
		}
		// eof from input means drop into stdin
		if err == io.EOF {
			reader = bufio.NewReader(os.Stdin)
			fmt.Println("*** Input now from stdin ***")
			usingStdin = true
		}
		// ignore blank lines and comments
		s = stripComments(s)
		if s == "" {
			continue
		}
		// it's a command, try it
		err = rs.dispatch(s)
		switch e := err.(type) {
		case exiter:
			if !verbose || usingStdin {
				fmt.Println(e.Error())
				e.Exit()
			}
			reader = bufio.NewReader(os.Stdin)
			fmt.Println("*** Exit requested while verbose: input now from stdin ***")
			usingStdin = true
		case error:
			fmt.Printf("%d)  -> Error: %s\n", linenumber, err)
		case nil:
		default:
		}
		if rs.vm != nil && (verbose || usingStdin) {
			rs.vm.Disassemble(rs.vm.IP())
		}
	}
}

func main() {
	// this needs to be filled in dynamically because the help function traverses
	// the commands list.
	h := commands["help"]
	h.handler = help
	commands["help"] = h

	var args struct {
		Binary  string `arg:"-b" help:"File to load as a chasm binary (*.chbin)."`
		Input   string `arg:"-i" help:"Input command file"`
		Verbose bool   `arg:"-v" help:"When executing input file, echo each line to the output; also causes errors to drop to repl."`
	}
	arg.MustParse(&args)
	var inf io.Reader
	if args.Input != "" {
		var err error
		inf, err = os.Open(args.Input)
		if err != nil {
			panic(err)
		}
	}

	rs := runtimeState{}
	if args.Binary != "" {
		err := rs.load(args.Binary)
		if err != nil {
			panic(err)
		}
	}

	rs.repl(inf, args.Verbose)
}
