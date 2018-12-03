package main

// we expect this to be invoked on OpcodeData
const tmplOpcodesPigeon = `
{{- define "BinOp"}}{ return newBinaryOpcode(vm.Op{{.Name}}, {{(index .Parms 0).Placeholder}}.(string)) }{{end}}
{{- define "PushB"}}{ return newPushB({{(index .Parms 0).Placeholder}}) }{{end}}
{{- define "PushT"}}{ return newPushTimestamp({{(index .Parms 0).Placeholder}}.(string)) }{{end}}
{{- define "CallOp"}}{ return newCallOpcode(vm.Op{{.Name}}, {{(index .Parms 0).Placeholder}}.(string)) }{{end}}
{{- define "DecoOp"}}{ return newDecoOpcode(vm.Op{{.Name}}, {{(index .Parms 0).Placeholder}}.(string), {{(index .Parms 1).Placeholder}}.(string)) }{{end}}
{{- range .ChasmOpcodes}}
{{- if eq 1 (len .Parms)}}
	{{- if eq "BinOp" (getparm . 0).PeggoTmpl}}
	/ "{{tolower .Name}}" _ {{(index .Parms 0).Placeholder}}:{{(index .Parms 0).PeggoParm}} {{template "BinOp" .}}
	{{- else if eq "PushB" (getparm . 0).PeggoTmpl}}
	/ "{{tolower .Name}}" _ {{(index .Parms 0).Placeholder}}:{{(index .Parms 0).PeggoParm}} {{template "PushB" .}}
	{{- else if eq "PushT" (getparm . 0).PeggoTmpl}}
	/ "{{tolower .Name}}" _ {{(index .Parms 0).Placeholder}}:{{(index .Parms 0).PeggoParm}} {{template "PushT" .}}
	{{- else if eq "CallOp" (getparm . 0).PeggoTmpl}}
	/ "{{tolower .Name}}" _ {{(index .Parms 0).Placeholder}}:{{(index .Parms 0).PeggoParm}} {{template "CallOp" .}}
	{{- end}}
{{- else if eq 2 (len .Parms)}}
    / "{{tolower .Name}}" _ {{(index .Parms 0).Placeholder}}:{{(index .Parms 0).PeggoParm}} _ {{(index .Parms 1).Placeholder}}:{{(index .Parms 1).PeggoParm}} {{template "DecoOp" .}}
{{- else}}
	/ "{{tolower .Name}}" { return newUnitaryOpcode(vm.Op{{.Name}}) }
{{- end}}
{{- end}}
`
