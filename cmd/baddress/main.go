package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/oneiro-ndev/commands/cmd/baddress/baddress"
)

// GenerateCmd handles args for generation
type GenerateCmd struct {
}

var args struct {
	Generate *GenerateCmd         `arg:"subcommand:generate" help:"automatically generate bad addresses"`
	Add      *baddress.BadAddress `arg:"subcommand:add" help:"manually add bad addresses"`
	Remove   *baddress.BadAddress `arg:"subcommand:remove" help:"manually remove bad address from the list"`
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
	sess, err := session.NewSession()
	check(err, "creating AWS session")
	ddb := dynamodb.New(sess)

	// then dispatch
	switch {
	case args.Generate != nil:
		fmt.Println("generate: unimplemented")
	case args.Add != nil:
		check(baddress.Add(ddb, *args.Add), "manually adding address")
	case args.Remove != nil:
		check(baddress.Remove(ddb, *args.Remove), "manually removing address")
	default:
		parser.WriteHelp(os.Stdout)
	}
}
