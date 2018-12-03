package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"

	"github.com/oneiro-ndev/ndaumath/pkg/words"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// the wordlist we want is private in another module. That's fine, all we need is its location
func getWords() []string {
	path := os.ExpandEnv("$GOPATH/src/github.com/oneiro-ndev/ndaumath/pkg/words/english.go")
	wfileb, err := ioutil.ReadFile(path)
	check(err)
	wfile := string(wfileb)

	re := regexp.MustCompile(`\s*"(\w+)",`)
	matches := re.FindAllStringSubmatch(wfile, -1)

	words := make([]string, 0, len(matches))
	for _, match := range matches {
		words = append(words, match[1])
	}
	return words
}

func twelve(s string) []string {
	out := make([]string, 12)
	for i := 0; i < 12; i++ {
		out[i] = s
	}
	return out
}

func main() {
	for _, word := range getWords() {
		bytes, err := words.ToBytes("en", twelve(word))
		if err == nil {
			fmt.Printf("[%s] * 12 -> %x\n", word, bytes)
			ekey, err := key.NewMaster(bytes)
			if err != nil {
				continue
			}
			addr, err := address.Generate(address.KindUser, ekey.PubKeyBytes())
			if err != nil {
				continue
			}
			fmt.Println("  (" + addr.String() + ")")
		}
	}
}
