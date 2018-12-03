package main

// we expect this to be invoked on OpcodeData
const tmplConstDef = `
// Code generated automatically by "make generate"; DO NOT EDIT.

package main

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
