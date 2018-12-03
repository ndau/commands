package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
)

func hdstr(k key.ExtendedKey) string {
	text, err := k.MarshalText()
	check(err)
	return string(text)
}

func cmdHDNew(cmd *cli.Cmd) {
	cmd.Action = func() {
		seed, err := key.GenerateSeed(key.RecommendedSeedLen)
		check(err)
		k, err := key.NewMaster(seed)
		check(err)
		fmt.Println(hdstr(*k))
	}
}

func cmdHDPublic(cmd *cli.Cmd) {
	cmd.Spec = getKeySpec("PVT")
	getKey := getKeyClosureHD(cmd, "PVT", "private key from which to make a public key")

	cmd.Action = func() {
		pvt := getKey()
		pub, err := pvt.Public()
		check(err)
		fmt.Println(hdstr(*pub))
	}
}

func cmdHDChild(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"%s PATH",
		getKeySpec(""),
	)

	getKey := getKeyClosureHD(cmd, "", "key from which to derive a child")

	pathS := cmd.StringArg("PATH", "", "derivation path for child key")
	cmd.Action = func() {
		key := getKey()
		key, err := key.DeriveFrom("/", *pathS)
		check(err)
		fmt.Println(hdstr(*key))
	}
}

func cmdHDConvert(cmd *cli.Cmd) {
	keyS := cmd.StringArg("KEY", "", "old-format key to convert")

	cmd.Action = func() {
		k, err := key.FromOldSerialization(*keyS)
		check(err)
		fmt.Println(hdstr(*k))
	}
}

func cmdHDAddr(cmd *cli.Cmd) {
	// mow.cli ensures with this that only one option is specified
	cmd.Spec = fmt.Sprintf(
		"%s %s",
		getKeySpec(""),
		getKindSpec(),
	)

	getKey := getKeyClosureHD(cmd, "", "get address from this key, converting to public as necessary")
	getKind := getKindClosure(cmd)

	cmd.Action = func() {
		key := getKey()
		kind := getKind()

		addr, err := address.Generate(kind, key.PubKeyBytes())
		check(err)
		fmt.Println(addr)
	}
}
