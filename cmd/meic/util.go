package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger logrus.FieldLogger

func check(err error, context string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, context+":")
		fmt.Fprintln(os.Stderr, err)
		if logger != nil {
			logger.WithError(err).WithField("context", context).Error("aborting")
		}
		os.Exit(1)
	}
}

func bail(err string, context ...interface{}) {
	if !strings.HasSuffix(err, "\n") {
		err += "\n"
	}
	fmt.Fprintf(os.Stderr, err, context...)
	os.Exit(1)
}

func input(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(prompt)
	inputline, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return strings.TrimSpace(inputline)
}

func getNested(dict map[string]interface{}, path ...string) (interface{}, error) {
	return getNestedInner(nil, dict, path)
}

func getNestedInner(breadcrumbs []string, dict map[string]interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return dict, nil
	}

	head := path[0]
	rest := path[1:]

	makeerr := func(msg string, context ...interface{}) error {
		prefix := " @ map"
		for _, b := range breadcrumbs {
			prefix += fmt.Sprintf("[%s]", b)
		}
		prefix += fmt.Sprintf("[%s]", head)
		return fmt.Errorf(msg+prefix, context...)
	}

	inner, ok := dict[head]
	if !ok {
		return nil, makeerr("item not found: %s", head)
	}

	if len(rest) == 0 {
		return inner, nil
	}

	imap, ok := inner.(map[string]interface{})
	if !ok {
		return nil, makeerr("unexpected item type: %T", inner)
	}

	return getNestedInner(append(breadcrumbs, head), imap, rest)
}
