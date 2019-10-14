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
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

// getRandomAccount randomly generates an account object
// it probably needs to be smarter than this
func getRandomAccount() backing.AccountData {
	const ticksPerDay = 24 * 60 * 60 * 1000000
	t, _ := types.TimestampFrom(time.Now())
	ad := backing.NewAccountData(t, types.Duration(rand.Intn(ticksPerDay*30)))
	// give it a balance between .1 and 100 ndau
	ad.Balance = types.Ndau((rand.Intn(1000) + 1) * 1000000)
	// set WAA to some time within 45 days
	ad.WeightedAverageAge = types.Duration(rand.Intn(ticksPerDay * 45))

	ad.LastEAIUpdate = t.Add(types.Duration(-rand.Intn(ticksPerDay * 3)))
	ad.LastWAAUpdate = t.Add(types.Duration(-rand.Intn(ticksPerDay * 10)))
	return ad
}

var predefined = predefinedConstants()

// parseInt parses an integer from a string. It is just like
// strconv.ParseInt(s, 0, bitSize) except that it can handle binary.
func parseInt(s string, bitSize int) (int64, error) {
	// remove embedded _ characters
	s = strings.Replace(s, "_", "", -1)
	if strings.HasPrefix(s, "0b") {
		return strconv.ParseInt(s[2:], 2, bitSize)
	}
	return strconv.ParseInt(s, 0, bitSize)
}

func parseValues(s string) ([]vm.Value, error) {
	result, err := Parse(fmt.Sprintf("parsing <<%s>>", s), []byte(s))
	if err != nil {
		return nil, err
	}
	return result.([]vm.Value), nil
}
