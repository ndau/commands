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
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/attic-labs/noms/go/spec"

	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/hash"
	nt "github.com/attic-labs/noms/go/types"
	cli "github.com/jawher/mow.cli"
	"github.com/pkg/errors"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		cli.Exit(1)
	}
}

func checkc(err error, context string) {
	check(errors.Wrap(err, context))
}

func bail(err string) {
	check(errors.New(err))
}

func main() {
	app := cli.App("nomsinfo", "get basic info about a noms db")
	app.LongDesc = strings.TrimSpace(`
For help specifying your datasets, see
https://github.com/attic-labs/noms/blob/master/doc/spelling.md
`)

	ds := app.StringArg("DATASET", "", "noms dataset")

	app.Action = func() {
		if ds == nil || *ds == "" {
			bail("ds must be set")
		}
		sp, err := spec.ForDataset(*ds)
		check(err)
		db := sp.GetDatabase()
		ds := sp.GetDataset()

		head, ok := ds.MaybeHeadRef()
		if !ok {
			bail("Dataset has no head ref")
		}

		fmt.Printf("%20s: %s\n", "apphash", apphash(head))
		fmt.Printf("%20s: %d\n", "noms height", head.Height())
		nodeheight, err := metanodeheight(db, head)
		if err == nil {
			fmt.Printf("%20s: %d\n", "node height", nodeheight)
		} else {
			fmt.Fprintf(os.Stderr, "%20s: %s\n", "bad metastate", err)
		}
	}
	app.Run(os.Args)
}

func apphash(ref nt.Ref) string {
	h := [hash.ByteLen]byte(ref.Hash())
	return hex.EncodeToString(h[:])
}

func valueAt(db datas.Database, ref nt.Ref) nt.Value {
	return ref.TargetValue(db).(nt.Struct).Get(datas.ValueField)
}

func metanodeheight(db datas.Database, ref nt.Ref) (uint64, error) {
	metastate, ok := valueAt(db, ref).(nt.Struct)
	if !ok {
		return 0, errors.New("expected metastate to be a nt.Struct")
	}
	heightv, ok := metastate.MaybeGet("Height")
	if !ok {
		return 0, errors.New("metastate did not have a .Height field")
	}
	heights, ok := heightv.(nt.String)
	if !ok {
		return 0, errors.New("expected .Height to be stored as a nt.String")
	}
	v, err := strconv.ParseUint(
		string(heights),
		36, 64,
	)
	if err != nil {
		return 0, errors.Wrap(err, "node height not a base36 string")
	}
	return v, nil
}
