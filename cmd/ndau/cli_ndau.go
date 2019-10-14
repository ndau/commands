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
