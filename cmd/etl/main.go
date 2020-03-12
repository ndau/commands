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
	"errors"
	"fmt"

	util "github.com/ndau/genesis/pkg/cli.util"
	"github.com/ndau/genesis/pkg/config"
	"github.com/ndau/genesis/pkg/etl"
)

func main() {
	ndauhome := util.GetNdauhome()
	path := config.DefaultConfigPath(ndauhome)
	var rows []etl.RawRow
	var err error
	err = config.WithConfig(path, func(conf *config.Config) error {
		rows, err = etl.Extract(conf)
		if err != nil {
			return err
		}
		duplicates := etl.DuplicateAddresses(rows)
		if len(duplicates) > 0 {
			fmt.Println("ERROR: duplicate addresses present:")
			for addr, rows := range duplicates {
				fmt.Printf("  %s:\n", addr)
				fmt.Printf("    ")
				for _, row := range rows {
					fmt.Printf("%d ", row)
				}
				fmt.Println()
			}
			return errors.New("duplicate addresses")
		}

		return etl.Load(conf, rows, ndauhome)
	})
	util.Check(err)
}
