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
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"
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

func main() {
	var args struct {
		Encode bool   `arg:"-E" help:"When set, encode a base10 number to base36"`
		Value  string `arg:"positional" help:"Decode or encode this number. If unset or -, read from stdin"`
	}
	args.Value = "-"
	arg.MustParse(&args)

	var value string
	if args.Value == "-" {
		data, err := ioutil.ReadAll(os.Stdin)
		check(err, "failed to read from stdin")
		value = string(data)
	} else {
		value = args.Value
	}
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "'\"")

	if args.Encode {
		n, err := strconv.ParseInt(value, 10, 64)
		check(err, "parsing %s as base10 integer", value)
		fmt.Println(strconv.FormatInt(n, 36))
	} else {
		n, err := strconv.ParseInt(value, 36, 64)
		check(err, "parsing %s as base36 integer", value)
		fmt.Println(n)
	}
}
