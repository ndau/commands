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
const tmplOpcodeDoc = `
# Opcodes for Chaincode

Generated automatically by "make generate"; DO NOT EDIT.

## Implemented and Enabled Opcodes

Value|Opcode|Meaning|Stack before|Instr.|Stack after
----|----|----|----|----|----
{{range .Enabled -}}
    {{- printf "0x%02x" .Value}}|
    {{- .Name}}{{if .Synonym}} ({{.Synonym}}){{end}}|
    {{- .Summary}}|
    {{- .Example.Pre}}|
    {{- .Example.Inst}}|
    {{- .Example.Post}}
{{end -}}

# Disabled Opcodes

Value|Opcode|Meaning|Stack before|Instr.|Stack after
----|----|----|----|----|----
{{range .Disabled -}}
    {{- printf "0x%02x" .Value}}|
    {{- .Name}}|
    {{- .Summary}}|
    {{- .Example.Pre}}|
    {{- .Example.Inst}}|
    {{- .Example.Post}}
{{else -}}
||There are no disabled opcodes at the moment.||
{{end -}}

`
