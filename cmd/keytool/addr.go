package main

import (
	"errors"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

func getKindSpec() string {
	// mow.cli ensures with this that only one option is specified
	return "[-k=<kind> | -a | -n | -e | -x | -b | -m]"
}

func getKindClosure(cmd *cli.Cmd) func() address.Kind {
	var (
		pkind  = cmd.StringOpt("k kind", string(address.KindUser), "manually specify address kind")
		kuser  = cmd.BoolOpt("a user", false, "address kind: user (default)")
		kndau  = cmd.BoolOpt("n ndau", false, "address kind: ndau")
		kendow = cmd.BoolOpt("e endowment", false, "address kind: endowment")
		kxchng = cmd.BoolOpt("x exchange", false, "address kind: exchange")
		kbpc   = cmd.BoolOpt("b bpc", false, "address kind: bpc")
		kmm    = cmd.BoolOpt("m market-maker", false, "address kind: market maker")
	)

	return func() address.Kind {
		kind := address.Kind(*pkind) // never nil dereference; defaults to user
		switch {
		case kuser != nil && *kuser:
			kind = address.KindUser
		case kndau != nil && *kndau:
			kind = address.KindNdau
		case kendow != nil && *kendow:
			kind = address.KindEndowment
		case kxchng != nil && *kxchng:
			kind = address.KindExchange
		case kbpc != nil && *kbpc:
			kind = address.KindBPC
		case kmm != nil && *kmm:
			kind = address.KindMarketMaker
		}

		if !address.IsValidKind(kind) {
			check(errors.New("invalid kind: " + string(kind)))
		}
		return kind
	}
}

func cmdAddr(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"%s %s",
		getKeySpec("PUB"),
		getKindSpec(),
	)

	getKey := getKeyClosure(cmd, "PUB", "get address from this key")
	getKind := getKindClosure(cmd)

	cmd.Action = func() {
		key := getKey()
		_, ok := key.(*signature.PublicKey)
		if !ok {
			check(errors.New("addresses can only be generated from public keys"))
		}

		kind := getKind()

		addr, err := address.Generate(kind, key.KeyBytes())
		check(err)
		fmt.Println(addr)
	}
}
