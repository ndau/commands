package main

import (
	"fmt"
	"os"
	"path"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/genesisfile"
	genesis "github.com/oneiro-ndev/chaos_genesis/pkg/genesis"
	"github.com/pkg/errors"
)

func ndauhome() string {
	ndauhome := os.ExpandEnv("$NDAUHOME")
	if len(ndauhome) == 0 {
		home := os.ExpandEnv("$HOME")
		ndauhome = path.Join(home, ".ndau")
	}
	return ndauhome
}

func nomsPath(ndauhome string) string {
	return path.Join(ndauhome, "chaos", "noms")
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(
			os.Stderr,
			err.Error(),
		)
		os.Exit(1)
	}
}

func main() {
	app := cli.App("chaos.genesis", "copy chaos genesisfile onto the chaos chain")

	var (
		verbose = app.BoolOpt("v verbose", false, "emit more detailed information")
		dryRun  = app.BoolOpt("d dry-run", false, "don't actually copy any data; just verify that the file format is valid")
		gfpath  = app.StringOpt(
			"g genesisfile",
			genesisfile.DefaultPath(ndauhome()),
			"path to genesisfile",
		)
		nomspath = app.StringOpt(
			"n noms",
			nomsPath(ndauhome()),
			"path to noms db",
		)
	)

	app.Action = func() {
		if verbose {
			fmt.Printf("%25s: %s\n", "genesisfile path", *gfpath)
		}

		gfile, err := genesisfile.Load(*gfpath)
		check(errors.Wrap(err, "loading genesis file"))
		check(genesis.Upload(gfile, *nomspath, *dryRun))
	}

	check(app.Run(os.Args))
}
