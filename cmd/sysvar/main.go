package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/tinylib/msgp/msgp"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

type args struct {
	Address address.Address `arg:"-a" help:"encode this ndau address"`
	Num     int64           `arg:"-i" help:"encode this integer"`
}

func (args) Description() string {
	return `
Encode specific types into appropriate formats for a SetSysvar tx
	`
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
	var args args
	arg.MustParse(&args)

	ea := address.Address{}
	if args.Address != ea {
		bytes, err := args.Address.MarshalMsg(nil)
		check(err, "msgp marshaling")
		fmt.Println(base64.StdEncoding.EncodeToString(bytes))
	} else {
		var bytes []byte
		bytes = msgp.Require(nil, msgp.Int64Size)
		bytes = msgp.AppendInt64(bytes, int64(args.Num))
		fmt.Println(base64.StdEncoding.EncodeToString(bytes))
	}
}
