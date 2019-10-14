package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// This is a formatter/indenter for .peggo files. It could use a lot of polish, but it's also
// adequate for our needs at the moment. Don't run it on new files you haven't committed
// to source control; it probably won't corrupt things but it's best to be safe.

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/kentquirk/rewriter/pkg/rewriter"
)

var section = `
	{
	fm := c.globalStore["functions"].(map[string]int)
		name := n.(string)
			ctr := c.globalStore["functionCounter"].(int)
  	fm[name] = ctr
	ctr   ++
	c.globalStore [ "functionCounter"] =ctr;
	return newFunctionDef(name, s)
	}
`

func formatSection(b []byte) ([]byte, error) {
	cmd := exec.Command("gofmt")
	cmd.Stdin = bytes.NewReader(b)
	var out bytes.Buffer
	out.WriteByte('@')
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// if we had an error, just return it unchanged
		return b, nil
	}
	out.WriteByte('@')
	return out.Bytes(), nil
}

func doNothing(b []byte) ([]byte, error) {
	// c := bytes.ToUpper(b)
	return b, nil
}

// we get back this nested mess of interface arrays
// so we have to unwind it and just keep the parts that actually text
func unwrap(v interface{}) string {
	s := ""
	switch v2 := v.(type) {
	case nil:
		return ""
	case []interface{}:
		for i := range v2 {
			s += unwrap(v2[i])
		}
		return s
	case []byte:
		return string(v2)
	case string:
		return v2
	default:
		return ""
	}
}

// untabify converts leading tabs to spaces and removes trailing space
func untabify(s string) string {
	s = strings.TrimRight(s, " \t\r\n")
	out := bytes.Buffer{}
outer:
	for i, c := range s {
		switch c {
		case ' ':
			out.WriteRune(c)
		case '\t':
			out.WriteRune(' ')
			for out.Len()%4 != 0 {
				out.WriteRune(' ')
			}
		default:
			out.WriteString(s[i:])
			break outer
		}
	}
	return out.String()
}

type indenter struct {
	pat          *regexp.Regexp
	deltacurrent int
	deltanew     int
}

var indenters = []indenter{
	indenter{regexp.MustCompile(`^} else.* {$`), -1, 0},
	// indenter{regexp.MustCompile(`^(/|\().*{$`), 0, 2},
	indenter{regexp.MustCompile(`@{$`), 0, 2},
	indenter{regexp.MustCompile(`}@$`), -1, -2},
	indenter{regexp.MustCompile(`^}$`), -1, -1},
	indenter{regexp.MustCompile("=$|<-$|\u2190$|\u27f5$|{$"), 0, 1},
	indenter{regexp.MustCompile(`^\)$`), 0, -1},
}

func reindent(s string, prev int, step int) (string, int) {
	s = strings.TrimSpace(s)
	current := prev
	newindent := current

	for _, i := range indenters {
		if i.pat.MatchString(s) {
			current += step * i.deltacurrent
			newindent += step * i.deltanew
			break // only do the first one we see
		}
	}

	if current < 0 {
		current = 0
	}
	if newindent < 0 {
		newindent = 0
	}
	s = strings.NewReplacer("@{", "{", "}@", "}").Replace(s)
	// fmt.Printf("%2d %2d %2d %s\n", prev, current, newindent, s)
	return strings.Repeat(" ", current) + s, newindent
}

type args struct {
	File    string `arg:"positional" help:"File to format; if not specified or just '-', reads from stdin and writes to stdout."`
	Tabsize int    `arg:"-s" help:"Change in indentation for each level [4]"`
}

func (args) Description() string {
	return `This program makes .peggo files reasonably pretty.`
}

func main() {
	a := args{
		Tabsize: 4,
		File:    "-",
	}

	arg.MustParse(&a)

	var infile io.ReadCloser = os.Stdin
	var outfile io.WriteCloser = os.Stdout
	if a.File != "-" {
		rwf, err := rewriter.New(a.File)
		if err != nil {
			log.Fatalf("Couldn't open %s: %s", a.File, err)
		}
		infile = rwf
		outfile = rwf
	}

	contents, err := ParseReader(a.File, infile)
	if err != nil {
		if infile.(rewriter.Aborter) != nil {
			infile.(rewriter.Aborter).Abort()
		}
		log.Fatalf("Couldn't parse %s: %s", a.File, err)
	}
	firstpass := unwrap(contents)
	linep := regexp.MustCompile("[ \t]*\n")
	lines := linep.Split(firstpass, -1)
	onelinep := regexp.MustCompile(`[ \t]*(.+?)[ \t]+({.+})$`)
	maxprefix := 0
	maxblock := 0
	indent := 0
	for _, l := range lines {
		l, indent = reindent(l, indent, a.Tabsize)
		if onelinep.MatchString(l) {
			indices := onelinep.FindStringSubmatchIndex(l)
			// indices[2] is the position of the first nonspace before the block, [3] is the last
			if indices[3] > maxprefix {
				maxprefix = indices[3]
			}
			// indices[4] is the position of the first char of the block, [5] is the last
			if indices[5]-indices[4] > maxblock {
				maxblock = indices[5] - indices[4]
			}
		}
	}
	_ = maxblock
	indent = 0
	// we want to suppress excess blank lines
	hadblank := false
	for _, l := range lines {
		l, indent = reindent(l, indent, a.Tabsize)
		if l == "" {
			hadblank = true
			continue
		} else {
			if hadblank {
				fmt.Fprintln(outfile)
				hadblank = false
			}
		}
		if onelinep.MatchString(l) {
			indices := onelinep.FindStringSubmatchIndex(l)
			fmt.Fprintf(outfile, "%*s %s\n", -maxprefix, l[indices[0]:indices[3]], l[indices[4]:])
		} else {
			fmt.Fprintf(outfile, "%s\n", l)
		}
	}
	infile.Close()
	outfile.Close()
}
