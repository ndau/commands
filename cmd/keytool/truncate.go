package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
)

func cmdTruncate(cmd *cli.Cmd) {
	cmd.Spec = getKeySpec("")

	getKey := getKeyClosure(cmd, "", "key to truncate")

	cmd.Action = func() {
		key := getKey()
		key.Truncate()
		keyB, err := key.MarshalText()
		check(err)
		fmt.Println(string(keyB))
	}
}
