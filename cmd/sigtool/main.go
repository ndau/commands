package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/oneiro-ndev/ndaumath/pkg/signature"

	arg "github.com/alexflint/go-arg"
)

type args struct {
	Generate   bool     `help:"generate a keypair from the input data"`
	Sign       bool     `help:"sign a block of data"`
	Verify     bool     `help:"verify a signature"`
	Hex        bool     `arg:"-x" help:"interpret input data as hex bytes"`
	Verbose    bool     `arg:"-v" help:"Be verbose"`
	OutputFile string   `arg:"-o" help:"Output filename"`
	Input      string   `arg:"-i" help:"Input filename"`
	Data       []string `arg:"positional"`
	Comment    string   `arg:"-c" help:"Comment for key files"`
	Keyfile    string   `arg:"-k" help:"Key filename"`
	Sigfile    string   `arg:"-s" help:"Signature filename"`
}

func (args) Description() string {
	return `
	Generates keypairs compatible with ndau, and signs blocks of data using those keys.

	Examples:
	signtool --generate bytestream -k keyfile
	# treats bytestream as entropy and generates a keypair, writing them to keyfile and keyfile.pub
	# default keyfile is "key"
	# if there is no input data at all, reads from the system entropy source

	signtool --sign -k keyfile bytestream
	# reads data from bytestream, hashes it, and signs it with the private key from keyfile;
	# sends the signature to outputfile

	signtool --verify -k keyfile -s sigfile bytestream
	# reads data from bytestream and verifies the signature (using keyfile.pub)

	bytestream can be command line arguments (strings or hex with -x), or you can use
	-i to read data from a named file; "-i -" reads from stdin.
	`
}

// reads an input stream and extracts content that it assumes is base64
func readAsBase64(in io.Reader) ([]byte, error) {
	dec := base64.NewDecoder(base64.StdEncoding, in)
	return ioutil.ReadAll(dec)
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

const keyType = "ed25519ndau"

// wrapHex encodes b as hex, wrapping at column w
func wrapHex(b []byte, w int) string {
	enc := hex.EncodeToString(b)
	out := ""
	for i := 0; len(enc) > 0; i += w {
		n := len(enc)
		if n > w {
			n = w
		}
		out += enc[0:n] + "\n"
		enc = enc[n:]
	}
	return out
}

func writePrivateKey(filename string, pvt signature.PrivateKey, note string) error {
	const (
		hdr      = "---- Begin Private Key ----"
		ftr      = "---- End Private Key ----"
		format   = hdr + "\nKey Type: %[1]s\nNote: %[3]s\n\n%[2]s\n" + ftr + "\n"
		maxWidth = 64
	)

	b, err := pvt.Marshal()
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	out := wrapHex(b, maxWidth)

	_, err = fmt.Fprintf(f, format, keyType, out, note)
	return err
}

func writePublicKey(filename string, pub signature.PublicKey, note string) error {
	b, err := pub.Marshal()
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := hex.EncodeToString(b)
	_, err = fmt.Fprintf(f, "%s %s %s\n", keyType, enc, note)
	return err
}
func readYubiPublicKey(filename string) (signature.PublicKey, error) {
	empty := signature.PublicKey{}
	f, err := os.Open(filename)
	if err != nil {
		return empty, err
	}
	defer f.Close()
	data, err := readAsBase64(f)
	if err != nil {
		return empty, err
	}
	key, err := signature.RawPublicKey(signature.Ed25519, data, nil)
	return *key, err
}

func readPublicKey(filename string) (signature.PublicKey, error) {
	// first try to read it as a yubikey file -- if it works, we're done
	key, err := readYubiPublicKey(filename)
	if err == nil {
		return key, nil
	}

	// ok, instead try to read our file
	sig := signature.PublicKey{}
	f, err := os.Open(filename)
	if err != nil {
		return sig, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	sp := strings.SplitN(string(data), " ", 3)
	if len(sp) < 2 || sp[0] != keyType {
		return sig, errors.New("Not an ndau public key file")
	}
	b, err := hex.DecodeString(sp[1])
	if err != nil {
		return sig, err
	}
	err = sig.Unmarshal(b)
	return sig, err
}

func readPrivateKey(filename string) (signature.PrivateKey, error) {
	sig := signature.PrivateKey{}
	f, err := os.Open(filename)
	if err != nil {
		return sig, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	sp := strings.Split(string(data), "\n")
	// TODO: Validate key type
	if len(sp) < 4 {
		return sig, errors.New("Not an ndau private key file")
	}
	keydata := strings.Join(sp[3:len(sp)-2], "")
	b, err := hex.DecodeString(keydata)
	if err != nil {
		return sig, err
	}
	err = sig.Unmarshal(b)
	return sig, err
}

func readYubiSignature(filename string) (signature.Signature, error) {
	empty := signature.Signature{}
	f, err := os.Open(filename)
	if err != nil {
		return empty, err
	}
	defer f.Close()
	data, err := readAsBase64(f)
	if err != nil {
		return empty, err
	}
	sig, err := signature.RawSignature(signature.Ed25519, data)
	return *sig, err
}

func readSignature(filename string) (signature.Signature, error) {
	// first try to read it as a yubikey file -- if it works, we're done
	sig, err := readYubiSignature(filename)
	if err == nil {
		return sig, nil
	}

	// ok, instead try to read our file
	f, err := os.Open(filename)
	if err != nil {
		return sig, err
	}
	defer f.Close()
	b, err := readAsHex(f)
	if err != nil {
		return sig, err
	}
	err = sig.Unmarshal(b)
	return sig, err
}

func main() {
	var args args
	args.Keyfile = "key"
	arg.MustParse(&args)

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

	var data []byte
	var err error
	if args.Hex {
		data, err = readAsHex(in)
	} else {
		data, err = ioutil.ReadAll(in)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if args.Verbose {
		fmt.Printf("input data:\n%s\n", hex.Dump(data))
	}

	switch {
	case args.Generate:
		if args.Verbose {
			fmt.Println("Generating...")
		}
		// we're creating a keypair
		var r io.Reader
		if len(data) > 0 {
			r = bytes.NewReader(data)
		}
		pub, pvt, err := signature.Generate(signature.Ed25519, r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s generating key\n", err)
			os.Exit(1)
		}
		err = writePublicKey(args.Keyfile+".pub", pub, args.Comment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s writing public key\n", err)
			os.Exit(1)
		}
		err = writePrivateKey(args.Keyfile, pvt, args.Comment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s writing private key\n", err)
			os.Exit(1)
		}
	case args.Sign:
		if args.Verbose {
			fmt.Println("Signing...")
		}
		// we're generating a signature so we need the private key
		if len(data) == 0 {
			fmt.Fprintln(os.Stderr, "we need data to sign")
			os.Exit(1)
		}

		pvt, err := readPrivateKey(args.Keyfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s reading private key '%s'\n", err, args.Keyfile)
			os.Exit(1)
		}
		sig := pvt.Sign(data)
		b, err := sig.Marshal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s marshalling signature\n", err)
			os.Exit(1)
		}

		f := os.Stdout
		if args.OutputFile != "" {
			f, err = os.Create(args.OutputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s creating output file\n", err)
				os.Exit(1)
			}
		}
		f.WriteString(wrapHex(b, 80))
		f.Close()

		pub, _ := readPublicKey(args.Keyfile + ".pub")
		fmt.Println(sig.Verify(data, pub))

	case args.Verify:
		if args.Verbose {
			fmt.Println("Verifying...")
		}
		if len(data) == 0 {
			fmt.Fprintln(os.Stderr, "we need data to verify a signature of it")
			os.Exit(1)
		}
		pub, err := readPublicKey(args.Keyfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s reading public key '%s'\n", err, args.Keyfile)
			os.Exit(1)
		}

		sig, err := readSignature(args.Sigfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s reading signature file '%s'\n", err, args.Keyfile)
			os.Exit(1)
		}
		if args.Verbose {
			fmt.Println("Sig:", sig)
		}
		if !sig.Verify(data, pub) {
			if args.Verbose {
				fmt.Println("Signature not verified!")
			}
			os.Exit(1)
		}
		if args.Verbose {
			fmt.Println("OK")
		}
	}
	os.Exit(0)

}
