package main

//go:generate gopherjs build --minify

// This is an experiment to see if gopherjs can reasonably generate js code from go source
// so that we can have a single-source solution for keys and addresses.
// Use "go generate" to build this.

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/miratronix/jopher"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

func main() {
	// validate accepts an ndau address (like those returned by generate) and returns
	// a promise that is resolved with true, or rejected with false.
	// function validate(addr: string): Promise
	js.Module.Get("exports").Set("validate", jopher.Promisify(validateWrapper))

	// generate creates an address of the appropriate kind (which must be one of
	// 'a', 'n', 'e', or 'x') and uses data (which must be at least 32 bytes long
	// and should be a public ndau key) to generate a new ndau key.
	// returns a Promise that resolves with the new key, or rejects with an error.
	// function generate(kind: string, data: string) : Promise
	js.Module.Get("exports").Set("generate", jopher.Promisify(generateWrapper))
}

func validateWrapper(addr string) (bool, error) {
	_, err := address.Validate(addr)
	return (err == nil), err
}

func generateWrapper(kind string, data string) (string, error) {
	k, err := address.NewKind(kind)
	if err != nil {
		return "", err
	}
	a, err := address.Generate(k, []byte(data))
	return a.String(), err
}
