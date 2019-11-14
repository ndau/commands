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
	"strings"

	"github.com/attic-labs/noms/go/datas"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
)

type state struct {
	db datas.Database
	ds datas.Dataset
	ms metast.Metastate
}

func (st state) state() *backing.State {
	return st.ms.ChildState.(*backing.State)
}

func (st state) summary() (out string) {
	var b strings.Builder
	defer func() {
		out = b.String()
	}()

	print := func(f string, as ...interface{}) {
		fmt.Fprintf(&b, f, as...)
		if f == "" || f[len(f)-1] != '\n' {
			b.WriteByte('\n')
		}
	}

	bs := st.state()
	if bs == nil {
		print("state is nil")
		return
	}

	print("state summary:")
	print("  %6d accounts", len(bs.Accounts))
	print("  %6d nodes", len(bs.Nodes))
	return
}
