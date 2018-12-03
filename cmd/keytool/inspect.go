package main

import (
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
		fmt.Printf("%10s: %x\n", "key", key.KeyBytes())
		fmt.Printf("%10s: %x\n", "extra", key.ExtraBytes())
	}
}
