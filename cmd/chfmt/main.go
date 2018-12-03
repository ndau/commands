package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
)

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

type line struct {
	keyword string
	args    string
	comment string
}

func tostring(i interface{}) string {
	switch v := i.(type) {
	case nil:
		return ""
	case string:
		return v
	case []byte:
		return string(v)
	case []interface{}:
		s := ""
		for _, v2 := range v {
			s += tostring(v2)
		}
		return s
	default:
		fmt.Printf("unknown value %#v\n", v)
		return "UNKNOWN"
	}
}

func newLine(k, a, c interface{}) line {
	return line{keyword: tostring(k), args: tostring(a), comment: tostring(c)}
}

type args struct {
	Input     string `arg:"positional" help:"Input file; if not specified, reads from stdin."`
	Indent    int    `arg:"-n" help:"Starting indent [0]"`
	Step      int    `arg:"-s" help:"Change in indentation for each level [4]"`
	Comment   int    `arg:"-c" help:"Leftmost column for inline comments [36]"`
	Overwrite bool   `arg:"-O" help:"Overwrite the input file with the formatted result. [false]"`
	Output    string `arg:"-o" help:"Output filename [stdout]"`
}

func (args) Description() string {
	return `This program makes .chasm files reasonably pretty. It:
	* aligns inline comments at the right
	* comments outside of functions are left-aligned
	* comments beginning with ;; are left-aligned to the current indent
	* handler, def, and if are indented by the stepsize
	* tabs are replaced by spaces and trailing spaces are trimmed
	`
}

func main() {
	a := args{
		Step:    4,
		Comment: 36,
	}

	arg.MustParse(&a)
	if a.Overwrite && a.Input != "" {
		a.Output = a.Input
	}

	name := "stdin"
	in := os.Stdin
	if a.Input != "" {
		name = a.Input
		f, err := os.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		in = f
	}

	var buf bytes.Buffer
	tee := io.TeeReader(in, &buf)

	alllines, err := ParseReader("", tee)
	if err != nil {
		log.Fatal(describeErrors(err, buf.String()))
	}

	// we successfully read the file, now close it in case
	// user wants to overwrite it
	in.Close()

	out := os.Stdout
	if a.Output != "" {
		f, err := os.Create(a.Output)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		out = f
	}

	indent := a.Indent
	newindent := indent
	for _, li := range toIfaceSlice(toIfaceSlice(alllines)[0]) {
		l, ok := li.(line)
		if !ok {
			fmt.Printf("not a line: %#v\n", li)
			continue
		}
		switch strings.ToLower(l.keyword) {
		case "handler", "def", "ifz", "ifnz":
			newindent += a.Step
		case "else":
			indent -= a.Step
		case "}", "enddef", "endif":
			newindent -= a.Step
			indent -= a.Step
		}

		// if we have args but no keyword, it's a constant and should be moved to the keyword field
		if l.keyword == "" && l.args != "" {
			l.keyword, l.args = l.args, l.keyword
		}

		// for debugging
		// fmt.Fprintf(out, "k:(%-8s) a:(%s)  c:(%s)\n", l.keyword, l.args, l.comment)

		// comment-only lines starting with ;; are always aligned to the current indent
		// rather than to the comment indent, as are lines that are not inside
		// a handler or function
		if l.keyword == "" && l.comment != "" &&
			(indent == a.Indent || strings.HasPrefix(l.comment, ";;")) {
			out.WriteString(strings.TrimRight(fmt.Sprintf("%*s%s", indent, "", l.comment), " ") + "\n")
		} else {
			code := fmt.Sprintf("%*s%s %s", indent, "", l.keyword, l.args)
			out.WriteString(strings.TrimRight(fmt.Sprintf("%-*s%s", a.Comment, code, l.comment), " ") + "\n")
		}

		indent = newindent
	}

}
