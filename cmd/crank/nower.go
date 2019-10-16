package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import "github.com/oneiro-ndev/chaincode/pkg/vm"

type nower struct{ now vm.Timestamp }

func (n nower) Now() (vm.Timestamp, error) {
	return n.now, nil
}

var _ vm.Nower = (*nower)(nil)
