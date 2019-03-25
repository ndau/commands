package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"

	cli "github.com/jawher/mow.cli"
	generator "github.com/oneiro-ndev/system_vars/pkg/genesis.generator"
	"github.com/oneiro-ndev/system_vars/pkg/genesisfile"
)

func ndauhome() string {
	ndauhome := os.ExpandEnv("$NDAUHOME")
	if len(ndauhome) == 0 {
		home := os.ExpandEnv("$HOME")
		ndauhome = path.Join(home, ".ndau")
	}
	return ndauhome
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
	app := cli.App("generate", "generate sysvar genesis file and associated data")

	var (
		verbose = app.BoolOpt("v verbose", false, "emit more detailed information")
		dryRun  = app.BoolOpt("d dry-run", false, "don't actually generate any data")
		gfpath  = app.StringOpt(
			"g genesisfile",
			genesisfile.DefaultPath(ndauhome()),
			"path to genesisfile",
		)
		afpath = app.StringOpt(
			"a associatedfile",
			generator.DefaultAssociated(ndauhome()),
			"path to genesisfile",
		)
		outpath = app.StringOpt("out", "", "put both generated files in this output directory")
	)

	// we have to specify the spec manually becuase inpath is incompatible
	// with the direct path specification options
	app.Spec = "[-v] [-d] [--out | [-g] [-a]]"

	app.Action = func() {
		if *verbose {
			if outpath != nil && len(*outpath) > 0 {
				fmt.Printf("%25s: %s\n", "output directory", *outpath)
			} else {
				fmt.Printf("%25s: %s\n", "genesisfile path", *gfpath)
				fmt.Printf("%25s: %s\n", "associatedfile path", *afpath)
			}
		}

		if !*dryRun {
			var bpc []byte
			var err error
			if outpath != nil && len(*outpath) > 0 {
				var gfilepath, asscpath string
				gfilepath, asscpath, err = generator.GenerateIn(*outpath)
				if *verbose {
					fmt.Printf("%25s: %s\n", "genesisfile path", gfilepath)
					fmt.Printf("%25s: %s\n", "associatedfile path", asscpath)
				}
			} else {
				err = generator.Generate(*gfpath, *afpath)
			}

			if bpc != nil {
				fmt.Printf(
					"%25s: %s\n",
					"b64 of bpc public key",
					base64.StdEncoding.EncodeToString(bpc),
				)
			}

			check(err)
		}
	}

	check(app.Run(os.Args))
}
