package main

// we expect this to be invoked on OpcodeData
const tmplOpcodesEnabled = `
// Code generated automatically by "make generate" -- DO NOT EDIT.

package vm

import "github.com/oneiro-ndev/ndaumath/pkg/bitset256"

// EnabledOpcodes is a bitset of the opcodes that are enabled -- only these opcodes will be
// permitted in scripts.
var EnabledOpcodes = bitset256.New(
{{- range .Enabled}}
	byte(Op{{.Name}}),
{{- end}}
)
`
