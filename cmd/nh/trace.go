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

	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
)

func (st state) trace(addr address.Address, out record) {
	var prevHeight uint64
	var prevAD backing.AccountData
	var existed bool
	err := metast.IterHistory(st.db, st.ds, &backing.State{}, func(stI metast.State, height uint64) error {
		defer func() { prevHeight = height }()
		out = out.Field("block", prevHeight)
		st := stI.(*backing.State)

		ad, exists := st.Accounts[addr.String()]
		defer func() {
			prevAD = ad
			existed = exists
		}()
		if !exists && existed {
			out.Emit("account created")
			return metast.StopIteration()
		}

		if prevHeight != 0 {
			var diff bool
			if ad.Balance != prevAD.Balance {
				// remember, we're iterating backwards, so 'prev' is really subsequent
				out = out.Field("balance", prevAD.Balance).
					Field("prev.balance", ad.Balance)
				diff = true
			}
			if len(ad.ValidationKeys) != len(prevAD.ValidationKeys) {
				out = out.Field("validation_keys.qty", len(prevAD.ValidationKeys)).
					Field("prev.validation_keys.qty", len(ad.ValidationKeys))
				diff = true
			}
			klen := len(ad.ValidationKeys)
			if len(prevAD.ValidationKeys) < klen {
				klen = len(prevAD.ValidationKeys)
			}
			for idx := 0; idx < klen; idx++ {
				ks := ad.ValidationKeys[idx].FullString()
				pks := prevAD.ValidationKeys[idx].FullString()
				if ks != pks {
					out = out.Field(fmt.Sprintf("validation_keys.%d", idx), pks).
						Field(fmt.Sprintf("prev.validation_keys.%d", idx), ks)
					diff = true
				}
			}
			if diff {
				out.Emit("account data change")
			}
		}

		return nil
	})
	check(err, "iterating noms history")
}
