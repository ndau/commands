package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/words"
)

const parentPath = "/44'/20036'/100"
const rootPath = "/"

func usage() {
	fmt.Fprintln(os.Stderr, "Example usage:")
	fmt.Fprintln(os.Stderr, "ndauArgs 17 >output.txt")
	fmt.Fprintln(os.Stderr, "This will prompt you for your 12 words, generate 17 ndau addresses,")
	fmt.Fprintln(os.Stderr, "and send the data to the file called 'output.txt'")
	os.Exit(1)
}

func main() {
	args := struct {
		Count int      `arg:"positional" help:"The number of keys to generate."`
		Words []string `arg:"positional" help:"The twelve words, in order; missing words will be prompted."`
	}{}
	arg.MustParse(&args)

	// we have to know how many to do
	if args.Count < 1 {
		usage()
	}

	reader := bufio.NewReader(os.Stdin)
	for len(args.Words) < 12 {
		fmt.Fprintf(os.Stderr, "Enter your passphrase (%d/12 so far): ", len(args.Words))
		inputline, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		w := strings.Fields(inputline)
		args.Words = append(args.Words, w...)
	}
	args.Words = args.Words[:12]

	rootbytes, err := words.ToBytes("en", args.Words)
	if err != nil {
		fmt.Fprintln(os.Stderr, "That list of words was not a valid passphrase. Please try again.")
		os.Exit(1)
	}
	rootkey, err := key.NewMaster(rootbytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s creating key\n", err.Error())
	}

	parentkey, err := rootkey.DeriveFrom(rootPath, parentPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s deriving parent key\n", err.Error())
	}
	for i := 1; i <= args.Count; i++ {
		childpath := fmt.Sprintf("%s/%d", parentPath, i)
		childkey, err := parentkey.DeriveFrom(parentPath, childpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error %s deriving child key %d\n", err.Error(), i)
		}
		a, err := address.Generate(address.KindUser, childkey.PubKeyBytes())
		fmt.Println(childpath, a)
	}
	fmt.Fprintf(os.Stderr, "%d addresses generated.\n", args.Count)
}
