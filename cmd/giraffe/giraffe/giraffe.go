package giraffe

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"io/ioutil"
	"os"
	"regexp"

	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/key"
	"github.com/ndau/ndaumath/pkg/words"
	"github.com/pkg/errors"
)

// GetEnWords gets the english-language wordlist for phrase generation
func GetEnWords() ([]string, error) {
	return GetWords(os.ExpandEnv("$GOPATH/src/github.com/ndau/ndaumath/pkg/words/english.go"))
}

// GetWords gets a wordlist for phrase generation
func GetWords(path string) ([]string, error) {
	wfileb, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading wordlist file")
	}
	wfile := string(wfileb)

	re := regexp.MustCompile(`\s*"(\w+)",`)
	matches := re.FindAllStringSubmatch(wfile, -1)

	words := make([]string, 0, len(matches))
	for _, match := range matches {
		words = append(words, match[1])
	}
	return words, nil
}

func twelve(s string) []string {
	out := make([]string, 12)
	for i := 0; i < 12; i++ {
		out[i] = s
	}
	return out
}

// Giraffe is an extended key, the word which generated it, and the resulting address
type Giraffe struct {
	Addr address.Address
	Key  *key.ExtendedKey
	Word string
}

// Giraffes returns and fills a channel of keypairs and addresses formed from
// a single word in the wordlist repeated 12 times.
func Giraffes(wordlist []string) <-chan Giraffe {
	out := make(chan Giraffe)

	go func() {
		defer close(out)
		for _, word := range wordlist {
			bytes, err := words.ToBytes("en", twelve(word))
			if err != nil {
				continue
			}
			ekey, err := key.NewMaster(bytes)
			if err != nil {
				continue
			}
			addr, err := address.Generate(address.KindUser, ekey.PubKeyBytes())
			if err != nil {
				continue
			}
			out <- Giraffe{
				Addr: addr,
				Key:  ekey,
				Word: word,
			}
		}
	}()

	return out
}
