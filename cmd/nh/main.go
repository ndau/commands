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
	"fmt"
	"os"

	arg "github.com/alexflint/go-arg"
	"github.com/attic-labs/noms/go/d"
	"github.com/attic-labs/noms/go/spec"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
)

func check(err error, context string, formatters ...interface{}) {
	if err != nil {
		if context[len(context)-1] == '\n' {
			context = context[:len(context)-1]
		}
		context += ": %s\n"
		formatters = append(formatters, err.Error())
		fmt.Fprintf(os.Stderr, context, formatters...)
		os.Exit(1)
	}
}

func bail(err string) {
	if err == "" {
		err = "fatal error"
	}
	if err[len(err)-1] != '\n' {
		err += "\n"
	}
	fmt.Fprint(os.Stderr, err)
	os.Exit(1)
}

type args struct {
	Path  string          `arg:"positional,required" help:"path to noms db"`
	App   string          `help:"app name (dataset name)"`
	Trace address.Address `help:"trace this account's history"`
	JSON  bool            `help:"when set, emit output as structured JSON"`
}

func main() {
	var args args
	args.App = "ndau"
	arg.MustParse(&args)

	sp, err := spec.ForDatabase(args.Path)
	check(err, "getting noms spec")

	var st state
	// we can fail to connect to noms for a variety of reasons, catch these here and report error
	// we use Try() because noms panics in various places
	err = d.Try(func() {
		st.db = sp.GetDatabase()
	})
	check(err, "connecting to noms db")

	st.ms.ChildState = new(backing.State)
	st.ms.ChildState.Init(st.db)
	st.ds, err = st.ms.Load(st.db, st.db.GetDataset(args.App), st.ms.ChildState)
	check(err, "loading existing state")

	// configure output
	var out record
	if args.JSON {
		out = new(JSONRecord).Writer(os.Stdout)
	} else {
		out = new(TextRecord).Writer(os.Stdout)
	}

	switch {
	case args.Trace.String() != "":
		st.trace(args.Trace, out)
	default:
		st.summary(out)
	}
}
