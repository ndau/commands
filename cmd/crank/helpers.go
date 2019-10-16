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
	"fmt"
	"io"
	"os"
)

// exiter
type exiter interface {
	Exit()
	Error() string
}

type exitError struct {
	code    int
	err     error
	context *runtimeState
}

func (e exitError) Exit() {
	os.Exit(e.code)
}

func (e exitError) Error() string {
	if e.err == nil {
		return ""
	}
	if e.context != nil {
		return fmt.Sprintf("%s\ncontext: %s\n%s", e.err.Error(), e.context.binary, e.context.vm)
	}
	return e.err.Error()
}

func newExitError(code int, err error, ctx *runtimeState) exitError {
	return exitError{code: code, err: err, context: ctx}
}

type outputRow struct {
	isError bool
	content []byte
}

type outputter struct {
	rows []outputRow
}

func newOutputter() *outputter {
	return &outputter{rows: make([]outputRow, 0)}
}

// Record records a row with an error flag
func (o *outputter) Record(isError bool, content []byte) {
	o.rows = append(o.rows, outputRow{isError: isError, content: content})
}

// Println records a string as a non-error, terminating it with \n
func (o *outputter) Println(content interface{}) {
	o.Record(false, []byte(fmt.Sprintln(content)))
}

// Printf records a formatted string as a non-error
func (o *outputter) Printf(format string, args ...interface{}) {
	o.Record(false, []byte(fmt.Sprintf(format, args...)))
}

// Errorln records a string as an error, terminating it with \n
func (o *outputter) Errorln(content interface{}) {
	o.Record(true, []byte(fmt.Sprintln(content)))
}

// Errorf records a formatted string as an error
func (o *outputter) Errorf(format string, args ...interface{}) {
	o.Record(true, []byte(fmt.Sprintf(format, args...)))
}

// Flush writes out the current error records and possibly the non-errors as well.
// It resets the buffers afterward.
func (o *outputter) Flush(out io.Writer, includeNonerrors bool) {
	for _, r := range o.rows {
		if r.isError || includeNonerrors {
			out.Write(r.content)
		}
	}
	o.rows = make([]outputRow, 0)
}
