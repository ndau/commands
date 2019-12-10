package main

import (
	"encoding/json"
	"fmt"
	"io"
)

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// a record is a composable field structure
// it's like a log message, but specific to our use case
type record interface {
	Field(string, interface{}) record
	Writer(io.Writer) record
	Emit(string)
}

type fields struct {
	titles []string
	values []interface{}
}

func (fs fields) Field(k string, v interface{}) fields {
	fs.titles = append(fs.titles, k)
	fs.values = append(fs.values, v)
	return fs
}

// A TextRecord can write itself as structured human-readable text
type TextRecord struct {
	fs     fields
	writer io.Writer
}

var _ record = (*TextRecord)(nil)

// Writer implements record by storing an output stream
func (tr TextRecord) Writer(w io.Writer) record {
	tr.writer = w
	return tr
}

// Field implements record by storing a field/value pair
func (tr TextRecord) Field(k string, v interface{}) record {
	tr.fs = tr.fs.Field(k, v)
	return tr
}

// Emit the header followed by the fields, nicely formatted
func (tr TextRecord) Emit(header string) {
	fmt.Fprintln(tr.writer, header+":")
	hmlen := 0
	for _, h := range tr.fs.titles {
		if len(h) > hmlen {
			hmlen = len(h)
		}
	}
	lfmt := fmt.Sprintf("  %%-%ds %%v\n", hmlen+1)
	for idx := range tr.fs.titles {
		fmt.Fprintf(tr.writer, lfmt, tr.fs.titles[idx]+":", tr.fs.values[idx])
	}
}

// A JSONRecord can write itself as structured JSON data
type JSONRecord struct {
	fs     fields
	writer io.Writer
}

var _ record = (*JSONRecord)(nil)

// Writer implements record by storing an output stream
func (tr JSONRecord) Writer(w io.Writer) record {
	tr.writer = w
	return tr
}

// Field implements record by storing a field/value pair
func (tr JSONRecord) Field(k string, v interface{}) record {
	tr.fs = tr.fs.Field(k, v)
	return tr
}

// Emit the header followed by the fields, nicely formatted
func (tr JSONRecord) Emit(string) {
	// omit the header
	inter := make(map[string]interface{})
	for idx := range tr.fs.titles {
		inter[tr.fs.titles[idx]] = tr.fs.values[idx]
	}
	data, err := json.Marshal(inter)
	check(err, "marshaling json record")
	data = append(data, '\n')
	_, err = tr.writer.Write(data)
	check(err, "sending json record to output stream")
}
