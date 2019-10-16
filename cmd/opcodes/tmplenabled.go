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
const tmplOpcodesEnabled = `
// Code generated automatically by "make generate" -- DO NOT EDIT.

package vm

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import "github.com/oneiro-ndev/ndaumath/pkg/bitset256"

// EnabledOpcodes is a bitset of the opcodes that are enabled -- only these opcodes will be
// permitted in scripts.
var EnabledOpcodes = bitset256.New(
{{- range .Enabled}}
	byte(Op{{.Name}}),
{{- end}}
)
`
