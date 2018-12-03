package main

// we expect this to be invoked on OpcodeData
const tmplOpcodesDef = `
// Code generated automatically by "make generate"; DO NOT EDIT.

package vm

// Opcode is a byte used to identify an opcode; we rely on it being a byte in some cases.
type Opcode byte

//go:generate stringer -trimprefix Op -type Opcode opcodes.go
//go:generate msgp -tests=0

// Opcodes
const (
{{range .EnabledWithSynonyms -}}
	Op{{.Name}} Opcode = {{printf "0x%02x" .Value}}
{{end}}
)
`
