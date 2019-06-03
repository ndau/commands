package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/version"
)

func bail(err string, context ...interface{}) {
	if !strings.HasSuffix(err, "\n") {
		err = err + "\n"
	}
	fmt.Fprintf(os.Stderr, err, context...)
	os.Exit(1)
}

func check(err error, context string, more ...interface{}) {
	if err != nil {
		bail(fmt.Sprintf("%s: %s", context, err.Error()), more...)
	}
}

const (
	exitAlways = iota
	exitIfErr
	exitIfNoErr
	exitNever
)

type mainargs struct {
	Net      string `arg:"-N" help:"net to configure: ('main', 'test', 'dev', 'local', or a URL)"`
	Node     int    `arg:"-n" help:"node number to which to connect"`
	Verbose  bool   `arg:"-v" help:"emit additional debug data"`
	Command  string `arg:"-c" help:"run this command"`
	CMode    int    `arg:"-C" help:"when to exit after running a command. 0 (default): always; 1: if no err; 2: if err; 3: never"`
	SysAccts string `arg:"--system-accts" help:"load system_accts.toml from this path"`
}

func (mainargs) Version() string {
	v, err := version.Get()
	if err != nil {
		v = err.Error()
	}
	return "ndsh " + v
}

func main() {
	args := mainargs{
		Net: "mainnet",
	}

	arg.MustParse(&args)

	client, err := getClient(args.Net, args.Node)
	check(err, "setting up connection to node")

	shell := NewShell(
		args.Verbose,
		client,
		Exit{},
		Help{},
		Recover{},
		ListAccounts{},
		Verbose{},
		Net{},
		Add{},
		Watch{},
		View{},
		New{},
		RecoverKeys{},
		Claim{},
		Tx{},
		ChangeValidation{},
		LoadSystemAccounts{},
		Sysvar{},
		ReleaseFromEndowment{},
		Transfer{},
		TransferAndLock{},
		Issue{},
		Version{},
		Summary{},
		CreateChild{},
		ChangeRecourse{},
		Delegate{},
		NominateNodeRewards{},
		RecordPrice{},
		CommandValidatorChange{},
		Lock{},
		Notify{},
		Stake{},
		SetRewardsDestination{},
		RegisterNode{},
		ClaimNodeReward{},
		CreditEAI{},
		Closeout{},
	)

	shell.VWrite("initialized shell...")

	if args.SysAccts != "" {
		err = shell.LoadSystemAccts(args.SysAccts)
		check(err, "loading system accounts")
	}

	if args.Command != "" {
		code := 0
		err := shell.Exec(args.Command)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			code = 1
		}
		switch args.CMode {
		case exitAlways:
			os.Exit(code)
		case exitIfErr:
			if err != nil {
				os.Exit(code)
			}
		case exitIfNoErr:
			if err == nil {
				os.Exit(code)
			}
		case exitNever:
			// nothing
		default:
			os.Exit(code)
		}
	}

	shell.VWrite("running shell...")
	shell.Run()
}
