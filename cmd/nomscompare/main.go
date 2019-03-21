package main

import (
	"fmt"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		cli.Exit(1)
	}
}

func checkc(err error, context string) {
	check(errors.Wrap(err, context))
}

func bail(err string) {
	check(errors.New(err))
}

func main() {
	app := cli.App("nomscompare", "compare two noms datasets")
	app.LongDesc = strings.TrimSpace(`
Recursively compares noms datasets.

For help specifying your datasets, see
https://github.com/attic-labs/noms/blob/master/doc/spelling.md
`)

	dsa := app.StringArg("DATASET_A", "", "first dataset")
	dsb := app.StringArg("DATASET_B", "", "second dataset")
	verbose := app.BoolOpt("v verbose", false, "emit additional output")
	height := app.IntOpt("h height", -1, "compare at a given noms height")

	app.Action = func() {
		log.SetLevel(log.InfoLevel)
		if *verbose {
			log.SetLevel(log.DebugLevel)
		}
		compareDS(*dsa, *dsb, *height)
		log.Debug("done")
	}
	app.Run(os.Args)
}
