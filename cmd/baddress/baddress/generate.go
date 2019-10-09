package baddress

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

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/oneiro-ndev/commands/cmd/giraffe/giraffe"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
)

var patterns = []string{
	"/",
	"/44'/20036'/100/%d",
}

// Generate a bunch of addresses known to be fishy, and submit each of them
// to the bad address DB.
func Generate(ddb *dynamodb.DynamoDB, verbose bool) error {
	words, err := giraffe.GetEnWords()
	if err != nil {
		return errors.Wrap(err, "getting wordlist")
	}
	for g := range giraffe.Giraffes(words) {
		for idx := 0; idx < 10; idx++ {
			for _, pattern := range patterns {
				if pattern == "" {
					pattern = "/"
				}
				var key = g.Key
				var err error
				if pattern != "/" {
					pattern = fmt.Sprintf(pattern, idx)
					key, err = key.DeriveFrom("/", pattern)
					if err != nil {
						return errors.Wrap(err, "deriving key")
					}
				}
				addr, err := address.Generate(address.KindUser, key.PubKeyBytes())
				if err != nil {
					return errors.Wrap(err, "generating address")
				}
				err = Add(
					ddb,
					BadAddress{
						Address: addr,
						Path:    pattern,
						Reason: fmt.Sprintf(
							"derives from 12-word phrase with identical words: %s",
							g.Word,
						),
					},
					false,
				)
				if err != nil {
					return errors.Wrap(err, "sending to DDB")
				}

			}
		}
	}
	return nil
}
