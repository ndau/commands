package main

import (
	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndau/pkg/version"
)

// Version displays version information
type Version struct{}

var _ Command = (*Version)(nil)

// Name implements Command
func (Version) Name() string { return "version" }

// Run implements Command
func (Version) Run(argvs []string, sh *Shell) (err error) {
	args := struct{}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	local, err := version.Get()
	if err != nil {
		sh.Write("getting local version: %s", err)
	}
	remote, _, err := tool.Version(sh.Node)
	if err != nil {
		sh.Write("getting remote version: %s", err)
	}

	sh.Write(" local: %s\nremote: %s", local, remote)

	return
}
