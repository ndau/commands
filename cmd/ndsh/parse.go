package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// parse a json-ish string into an appropriate value
//
// This is a bit more forgiving than actual json, because it acknowledges that
// humans tend to i.e. leave quotes off of strings.
func parseJSON(s string) (value interface{}, err error) {
	var autoquote bool
	if s == "null" ||
		(s[0] == '"' && s[len(s)-1] == '"') ||
		(s[0] == '[' && s[len(s)-1] == ']') ||
		(s[0] == '{' && s[len(s)-1] == '}') {
		autoquote = false
	} else if _, err := strconv.ParseFloat(s, 64); err == nil {
		autoquote = false
	} else {
		autoquote = true
	}
	if autoquote {
		s = fmt.Sprintf("\"%s\"", s)
	}

	// we have to json-unmarshal the value in order to ensure that
	// we set the right datatype
	err = json.Unmarshal([]byte(s), &value)
	if err != nil {
		err = errors.Wrap(err, "interpreting value as JSON")
	}
	return
}
