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

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/oneiro-ndev/commands/cmd/baddress/baddress"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

// GenerateCmd handles args for generation
type GenerateCmd struct {
}

// RemoveCmd handles args for removing addresses
type RemoveCmd struct {
	Address address.Address `arg:"positional,required" help:"remove this address"`
}

// CheckCmd handles args for checking
type CheckCmd struct {
	Address  address.Address `arg:"positional,required" help:"check this address"`
	ExitCode bool            `arg:"-e,--exit-code" help:"return with an exit code of 2 if the address is in the db"`
}

var args struct {
	Generate *GenerateCmd         `arg:"subcommand:generate" help:"automatically generate bad addresses"`
	Add      *baddress.BadAddress `arg:"subcommand:add" help:"manually add bad addresses"`
	Remove   *RemoveCmd           `arg:"subcommand:remove" help:"manually remove bad address from the list"`
	Check    *CheckCmd            `arg:"subcommand:check" help:"check whether an address is valid or not"`
	Verbose  bool                 `arg:"-v" help:"emit additional information"`
}

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
	// first parse the CLI args
	parser := arg.MustParse(&args)

	// then get the AWS session
	sess, err := session.NewSession(&aws.Config{Region: aws.String(baddress.Region)})
	check(err, "creating AWS session")
	ddb := dynamodb.New(sess)

	// then dispatch
	switch {
	case args.Generate != nil:
		check(baddress.Generate(ddb, args.Verbose), "generating bad addresses")
	case args.Add != nil:
		check(baddress.Add(ddb, *args.Add, args.Verbose), "manually adding address")
	case args.Remove != nil:
		check(baddress.Remove(ddb, args.Remove.Address, args.Verbose), "manually removing address")
	case args.Check != nil:
		exists, err := baddress.Check(ddb, args.Check.Address)
		check(err, "checking whether %s is in bad address db", args.Check.Address)
		status := "does not exist"
		if exists {
			status = "exists"
		}
		fmt.Println(args.Check.Address, status)
		if args.Check.ExitCode && exists {
			os.Exit(2)
		}
	default:
		parser.WriteHelp(os.Stdout)
	}
}
