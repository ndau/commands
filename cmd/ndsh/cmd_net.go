package main

import (
	"github.com/alexflint/go-arg"
	tmclient "github.com/tendermint/tendermint/rpc/client"
)

// Net prints the net currently connected to, or updates it
type Net struct{}

var _ Command = (*Net)(nil)

// Name implements Command
func (Net) Name() string { return "net" }

// Run implements Command
func (Net) Run(argvs []string, sh *Shell) (err error) {
	args := struct {
		Set string `help:"switch networks to this network. WARNING: this can cause inconsistent state, only do this if you know what you're doing."`
		Num int    `help:"node number to use when switching networks"`
	}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	if args.Set != "" {
		var client tmclient.ABCIClient
		client, err = getClient(args.Set, args.Num)
		if err != nil {
			return
		}
		sh.Node = client
	}
	// ClientURL gets updated as a side-effect of getClient
	sh.Write(ClientURL.String())
	return
}
