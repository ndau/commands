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

	"github.com/oneiro-ndev/commands/cmd/giraffe/giraffe"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	words, err := giraffe.GetEnWords()
	check(err)
	for g := range giraffe.Giraffes(words) {
		fmt.Printf("[%s] * 12 -> %s\n", g.Word, g.Addr)
	}
}
