package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

type args struct {
	Address   []address.Address `arg:"-a,separate" help:"encode this ndau address"`
	Duration  []math.Duration   `arg:"-d,separate" help:"encode this duration"`
	Int64     []int64           `arg:"-i,separate" help:"encode this signed integer"`
	Ndau      []math.Ndau       `arg:"-n,separate" help:"encode this qty of ndau"`
	Uint64    []uint64          `arg:"-u,separate" help:"encode this unsigned integer"`
	StringArg []string          `arg:"-s,separate" help:"encode this string"`
	B64Bytes  []string          `arg:"-b,separate" help:"encode this base64-encoded array of bytes"`
}

func (args) Description() string {
	return `
Encode specific types into appropriate formats for a SetSysvar tx

Each flag may be set multiple times.
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

func output(obj interface{}, bytes []byte) {
	fmt.Printf("%-50v %s\n", obj, base64.StdEncoding.EncodeToString(bytes))
}

func main() {
	var args args
	arg.MustParse(&args)

	for _, v := range args.Address {
		bytes, err := v.MarshalMsg(nil)
		check(err, "msgp marshaling address")
		output(v.String(), bytes)
	}
	for _, v := range args.Duration {
		bytes, err := v.MarshalMsg(nil)
		check(err, "msgp marshaling duration")
		output(v, bytes)
	}
	for _, v := range args.Int64 {
		bytes, err := wkt.Int64(v).MarshalMsg(nil)
		check(err, "msgp marshaling int64")
		output(v, bytes)
	}
	for _, v := range args.Ndau {
		bytes, err := v.MarshalMsg(nil)
		check(err, "msgp marshaling ndau")
		output(v, bytes)
	}
	for _, v := range args.Uint64 {
		bytes, err := wkt.Uint64(v).MarshalMsg(nil)
		check(err, "msgp marshaling uint64")
		output(v, bytes)
	}
	for _, v := range args.StringArg {
		bytes, err := wkt.String(v).MarshalMsg(nil)
		check(err, "msgp marshaling string")
<<<<<<< Updated upstream
=======
		output(v, bytes)
	}
	for _, v := range args.B64Bytes {
		b, err := base64.StdEncoding.DecodeString(v)
		check(err, "decoding base64 input")
		bytes, err := wkt.Bytes(b).MarshalMsg(nil)
		check(err, "msgp marshaling uint64")
>>>>>>> Stashed changes
		output(v, bytes)
	}
}
