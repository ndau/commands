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
	"github.com/attic-labs/noms/go/datas"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

type state struct {
	db datas.Database
	ds datas.Dataset
	ms metast.Metastate
}

func (st state) state() *backing.State {
	return st.ms.ChildState.(*backing.State)
}

func (st state) summary(out record) {
	bs := st.state()
	if bs == nil {
		out.Emit("state is nil")
		return
	}

	out.Field("block height", st.ms.Height).Field("accounts", len(bs.Accounts)).Field("nodes", len(bs.Nodes)).Emit("state summary")
}

func (st state) trace(addr address.Address, out record) {

}
