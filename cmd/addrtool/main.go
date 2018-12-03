package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

type args struct {
	Generate bool     `arg:"-g" help:"generate an address from the input data"`
	Hex      bool     `arg:"-x" help:"interpret input data as hex bytes"`
	Verbose  bool     `arg:"-v" help:"Be verbose"`
	Input    string   `arg:"-i" help:"Input filename"`
	Kind     string   `arg:"-k" help:"Kind of address to create -- a, n, e, or x"`
	Data     []string `arg:"positional"`
}

func (args) Description() string {
	return `
	Generates or validates ndau account addresses.

	Examples:
		addrtool ndaei7fgutsm97h8hxfhkvbcdruqxu9p7narzerrh88gihhq
		# valid, no output, errcode 0

		addrtool ndbad
		# invalid, err output, errcode >0

		addrtool --verbose --input FILENAME
		# parses the contents of FILENAME; any strings in it that appear to be ndau addresses
		# (those that start with 'nd') are checked for validity. Results are printed for
		# both valid and invalid values.
		# If the verbose flag is not specified, only invalid values are printed to stderr, and
		# the errorcode is set to 1 if any invalid values are found.
		# If FILENAME is - reads stdin

		addrtool --verbose --generate TextOfMyPublicKey
		# input data:
		# 00000000  54 65 78 74 4f 66 4d 79  50 75 62 6c 69 63 4b 65  |TextOfMyPublicKe|
		# 00000010  79                                                |y|
		#
		# ndaar2qpvnyz58mucwx7wux7wtt372djicpzg9diebfsgmwf

		addrtool --generate -i key.pub
		# if key.pub is an ndau public key, this uses it to generate a proper ndau public address from it

		head -c 100 /dev/urandom |addrtool --generate --input -
		# Reads 100 random bytes from /dev/urandom and pipes them to addrtool to generate a new key.

		head -c 20 /dev/urandom |od -t x1 -A n |./addrtool -g -x -v -i -
		# Reads 20 random bytes from /dev/urandom, converts them to hex bytes, and instructs
		# addrtool to parse them as hex bytes and generate a new key (verbosely).
`
}

// reads an input stream and extracts anything that looks like ndau values
func readAddresses(in io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanWords)
	// extract the possible addresses
	addrs := []string{}
	for scanner.Scan() {
		s := scanner.Text()
		if strings.HasPrefix(s, "nd") {
			addrs = append(addrs, s)
		}
	}
	return addrs, scanner.Err()
}

// reads an input stream and extracts anything that looks like pairs of hex characters (bytes)
func readAsHex(in io.Reader) ([]byte, error) {
	all, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	pat := regexp.MustCompile("[0-9a-fA-F]{2}")
	// extract the possible hex values
	hexes := pat.FindAllString(string(all), -1)
	var output []byte
	// now convert them
	for _, h := range hexes {
		b, _ := strconv.ParseUint(h, 16, 8)
		output = append(output, byte(b))
	}
	return output, nil
}

func main() {
	var args args
	args.Kind = "a"
	arg.MustParse(&args)

	if !address.IsValidKind(address.Kind(args.Kind)) {
		fmt.Fprintf(os.Stderr, "%s is not a valid Kind\n", args.Kind)
		os.Exit(1)
	}

	// figure out where we get our input stream from
	var in io.Reader
	in = strings.NewReader(strings.Join(args.Data, " "))
	if args.Input != "" {
		if args.Input == "-" {
			in = os.Stdin
		} else {
			f, err := os.Open(args.Input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			defer f.Close()
			in = f
		}
	}

	if !args.Generate {
		// are we validating?
		errcode := 0
		addrs, err := readAddresses(in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if len(addrs) == 0 {
			fmt.Fprintln(os.Stderr, "no data that appeared to be ndau addresses")
			os.Exit(1)
		}

		for _, addr := range addrs {
			a, v := address.Validate(addr)
			if v == nil {
				if args.Verbose {
					fmt.Printf("%s: valid\n", a)
				}
			} else {
				errcode = 1
				if args.Verbose {
					fmt.Printf("%s: %s\n", a, v)
				} else {
					fmt.Fprintf(os.Stderr, "%s: %s\n", a, v)
				}
			}
		}
		os.Exit(errcode)
	}

	// we're generating an address
	var data []byte
	var err error
	if args.Hex {
		data, err = readAsHex(in)
	} else {
		// read data from the file
		data, err = ioutil.ReadAll(in)
		// but take a look to see if the data looks like a public key
		// if so, treat the key as the data stream
		sp := bytes.Split(data, []byte{' '})
		if len(sp) == 3 && string(sp[0]) == "ed25519ndau" {
			// it was a key file, so treat it as a stream of hex
			h, err := readAsHex(bytes.NewReader(sp[1]))
			if err == nil {
				data = h
			}
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if args.Verbose {
		fmt.Printf("input data:\n%s\n", hex.Dump(data))
	}
	a, err := address.Generate(address.Kind(args.Kind), data)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	fmt.Println(a)
}
