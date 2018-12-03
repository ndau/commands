package main

// we expect this to be invoked on OpcodeData
const tmplOpcodesMiniAsm = `
// Code generated automatically by "make generate"; DO NOT EDIT.

package vm

// these are the opcodes supported by mini-asm
var opcodeMap = map[string]Opcode{
{{range .EnabledWithSynonyms -}}
	"{{tolower .Name}}": Op{{.Name}},
{{end}}
}
`
