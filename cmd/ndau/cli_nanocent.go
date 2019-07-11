package main

import (
	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/pkg/errors"
)

func getNanocentSpec() string {
	return "(USD | --nanocents=<NANOCENTS>)"
}

func getNanocentClosure(cmd *cli.Cmd) func() pricecurve.Nanocent {
	var (
		usd       = cmd.StringArg("USD", "", "US Dollars")
		nanocents = cmd.IntOpt("nanocents", 0, "Integer quantity of nanocents. Allows sub-cent precision.")
	)

	return func() pricecurve.Nanocent {
		if nanocents != nil && *nanocents != 0 {
			return pricecurve.Nanocent(*nanocents)
		}
		if usd != nil && *usd != "" {
			nc, err := pricecurve.ParseDollars(*usd)
			orQuit(errors.Wrap(err, "parsing usd"))
			return nc
		}
		orQuit(errors.New("usd and nanocent not set"))
		return pricecurve.Nanocent(0)
	}
}
