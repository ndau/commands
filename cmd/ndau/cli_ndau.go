package main

import (
	"strconv"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

func getNdauSpec() string {
	return "(QTY | --napu=<NAPU>)"
}

func getNdauClosure(cmd *cli.Cmd) func() math.Ndau {
	var (
		qty  = cmd.StringArg("QTY", "", "qty of ndau")
		napu = cmd.IntOpt("napu", 0, "Qty of ndau expressed in terms of napu")
	)

	return func() math.Ndau {
		if napu != nil && *napu != 0 {
			return math.Ndau(*napu)
		}
		if qty != nil && *qty != "" {
			nf, err := strconv.ParseFloat(*qty, 64)
			orQuit(errors.Wrap(err, "parsing ndau"))
			return math.Ndau(nf * constants.QuantaPerUnit)
		}
		orQuit(errors.New("qty ndau not set"))
		return math.Ndau(0)
	}
}
