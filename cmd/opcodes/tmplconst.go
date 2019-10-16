package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// we expect this to be invoked on OpcodeData
const tmplConstDef = `
// Code generated automatically by "make generate"; DO NOT EDIT.

package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// Predefined constants available to chasm programs.

func predefinedConstants() map[string]string {
	k := map[string]string{
{{range . -}}
		"{{.Name}}": "{{printf "%d" .Value}}",
{{end}}
	}
	return k
}
`
