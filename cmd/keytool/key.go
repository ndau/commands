package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

// ktype should always be "", "PUB", or "PVT"
func keytype(ktype string) string {
	return strings.ToUpper(strings.TrimSpace(ktype)) + "KEY"
}

func getKeySpec(ktype string) string {
	return fmt.Sprintf("(%s | --stdin)", keytype(ktype))
}

func getKeyClosure(cmd *cli.Cmd, ktype string, desc string) func() signature.Key {
	key := cmd.StringArg(keytype(ktype), "", desc)
	stdin := cmd.BoolOpt("S stdin", false, "if set, read the key from stdin")

	return func() signature.Key {
		var keys string
		if stdin != nil && *stdin {
			in := bufio.NewScanner(os.Stdin)
			if !in.Scan() {
				check(errors.New("stdin selected but empty"))
			}
			check(in.Err())
			keys = in.Text()
		} else if key != nil && len(*key) > 0 {
			keys = *key
		} else {
			check(errors.New("no or multiple keys input--this should be unreachable"))
		}

		k, err := signature.ParseKey(keys)
		check(err)
		return k
	}
}

func getKeyClosureHD(cmd *cli.Cmd, ktype string, desc string) func() *key.ExtendedKey {
	getKey := getKeyClosure(cmd, ktype, desc)
	return func() *key.ExtendedKey {
		k := getKey()
		ek, err := key.FromSignatureKey(k)
		check(err)
		return ek
	}
}
