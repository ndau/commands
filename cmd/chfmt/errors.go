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
	"regexp"
	"strings"
)

// This file provides helpers to make errors found during parsing more palatable.

// ErrorPosition defines the raw error position data.
type ErrorPosition struct {
	name   string
	line   int
	col    int
	offset int
}

// ErrorPositioner is an interface that can be used to tell if an error provides
// position data in the source file.
type ErrorPositioner interface {
	ErrorPos() ErrorPosition
}

func (p *parserError) ErrorPos() ErrorPosition {
	return ErrorPosition{
		name:   p.prefix,
		line:   p.pos.line,
		col:    p.pos.col,
		offset: p.pos.offset,
	}
}

func describeError(err error, source, filename string) string {
	if e, ok := err.(ErrorPositioner); ok {
		lines := strings.Split(source, "\n")
		ep := e.ErrorPos()
		// get the line with the error
		line := lines[ep.line-1]
		// now create a second line with the same whitespace prefix, plus replace all
		// the non-whitespace chars with a space, then add a caret (^) to point to the error
		// c := ep.col
		if ep.col >= len(line) {
			ep.col = len(line) - 1
		}
		caretline := line[:ep.col]
		nonspace := regexp.MustCompile("[^ \t]")
		caretline = nonspace.ReplaceAllString(caretline, " ") + "^"
		return fmt.Sprintf("[%s] %s\n%4d: %s\n     %s %v\n", filename, err.Error(), ep.line, line, caretline, ep)
	}
	fmt.Printf("NOT ErrorPositioner: %#v\n", err)
	return err.Error()
}

func describeErrors(err error, source, filename string) string {
	if el, ok := err.(errList); ok {
		s := ""
		for _, e := range el {
			s += describeError(e, source, filename)
		}
		return s
	}
	fmt.Printf("NOT errList: %#v\n", err)
	return describeError(err, source, filename)
}
