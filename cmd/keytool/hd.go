package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
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

func cmdHDRaw(cmd *cli.Cmd) {
	cmd.Command("public", "transform a raw secp256k1 public key into ndau format", cmdHDRawPublic)
	cmd.Command("private", "transform a raw secp256k1 private key into ndau format", cmdHDRawPrivate)
	cmd.Command("signature", "transform a raw secp256k1 signature into ndau format", cmdHDRawSig)
}

func cmdHDRawPublic(cmd *cli.Cmd) {
	cmd.Spec = getDataSpec(true)

	getData := getDataClosure(cmd, true)

	cmd.Action = func() {
		data := getData()

		key, err := signature.RawPublicKey(signature.Secp256k1, data, nil)
		check(err)

		data, err = key.MarshalText()
		check(err)
		fmt.Println(string(data))
	}
}

func cmdHDRawPrivate(cmd *cli.Cmd) {
	cmd.Spec = getDataSpec(true)

	getData := getDataClosure(cmd, true)

	cmd.Action = func() {
		data := getData()

		base := make([]byte, 32)
		extra := make([]byte, 40)

		copy(base, data[0:32])
		if len(data) == 72 {
			copy(extra, data[32:72])
		}

		key, err := signature.RawPrivateKey(signature.Secp256k1, base, extra)

		check(err)

		data, err = key.MarshalText()
		check(err)
		fmt.Println(string(data))
	}
}

func cmdHDRawSig(cmd *cli.Cmd) {
	cmd.Spec = getDataSpec(true)

	getData := getDataClosure(cmd, true)

	cmd.Action = func() {
		data := getData()

		sig, err := signature.RawSignature(signature.Secp256k1, data)
		check(err)

		data, err = sig.MarshalText()
		check(err)
		fmt.Println(string(data))
	}
}
