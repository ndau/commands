package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func checkw(err error, context string) {
	check(errors.Wrap(err, context))
}

func main() {
	shell := NewShell(
		Exit{},
	)
	shell.Run()
}
