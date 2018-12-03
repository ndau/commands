package main

import (
	"fmt"
	"os"

	arg "github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/txmobile/generator"
	"github.com/pkg/errors"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

type args struct {
	Exclude []string `arg:"positional" help:"tx names to exclude"`
}

func main() {
	var args args
	arg.MustParse(&args)

	ast, err := generator.ParseTransactions()
	check(err)
	fmt.Println("parsed transactions")

	txIDs := generator.FindDefinition(ast, generator.TxIDs)
	if txIDs == nil {
		check(errors.New("TxIDs not found"))
	}

	fmt.Printf(
		"found TxIDs [%d:%d]: %s\n",
		txIDs.Definition.Pos(), txIDs.Definition.End(),
		txIDs.DefinitionType,
	)

	// emit the AST of TxIDs in a pretty-ish way
	// check(txIDs.Write(os.Stdout))

	txNames, err := generator.GetTxNames(txIDs.Definition)
	check(err)

	tmpl, err := generator.ParseTemplate()
	check(err)
	fmt.Println("successfully parsed template")

	fmt.Printf("Found %d names:\n", len(txNames))

nameloop:
	for _, n := range txNames {
		fmt.Printf("%25s: ", n)
		for _, x := range args.Exclude {
			if n == x {
				fmt.Println("EXCLUDED")
				continue nameloop
			}
		}
		def := generator.FindDefinition(ast, n)
		found := "not found"
		if def != nil {
			found = fmt.Sprintf(
				"found at [%d:%d]",
				def.Definition.Pos(),
				def.Definition.End(),
			)
		}
		fmt.Print(found)

		if def != nil {
			// emit AST of typedef
			// f, err := os.Create(fmt.Sprintf("%s.ast", n))
			// check(err)
			// check(def.Write(f))
			// f.Close()

			fmt.Print("...")
			transaction, err := generator.ParseTransaction(n, def.Definition)
			err = errors.Wrap(err, fmt.Sprintf("parsing %s tx", n))
			check(err)
			check(generator.ApplyTemplate(tmpl, transaction))
			fmt.Println(" DONE")
		}
	}
}
