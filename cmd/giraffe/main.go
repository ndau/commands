package main

import (
	"fmt"
	"os"

	"github.com/oneiro-ndev/commands/cmd/giraffe/giraffe"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	words, err := giraffe.GetEnWords()
	check(err)
	for g := range giraffe.Giraffes(words) {
		fmt.Printf("[%s] * 12 -> %s\n", g.Word, g.Addr)
	}
}
