package main

import (
	"encoding/base64"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

func cmdInspect(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"%s",
		getKeySpec(""),
	)

	getKey := getKeyClosure(cmd, "", "key to be inspected")

	cmd.Action = func() {
		key := getKey()

		ktype := "public"
		if signature.IsPrivate(key) {
			ktype = "private"
		}

		fmt.Printf("%10s: %s\n", "type", ktype)
		fmt.Printf("%10s: %s\n", "algorithm", signature.NameOf(key.Algorithm()))
		fmt.Printf("%10s: %x\n", "key hex", key.KeyBytes())
		fmt.Printf("%10s: %s\n", "key b64", base64.StdEncoding.EncodeToString(key.KeyBytes()))

		extraBytes := key.ExtraBytes()

		if len(extraBytes) != 0 {
			fmt.Printf("\n%20s: %x\n", "extended key depth", extraBytes[0])
			fmt.Printf("%20s: %x\n", "parent fingerprint", extraBytes[1:4])
			fmt.Printf("%20s: %x\n", "child number", extraBytes[4:8])
			fmt.Printf("%20s: %x\n", "chain code", extraBytes[8:40])
		}
	}
}
