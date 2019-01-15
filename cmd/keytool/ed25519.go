package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

func writeKey(w io.Writer, k signature.Key, verbose *bool, vprefix string) error {
	text, err := k.MarshalText()
	if err != nil {
		return err
	}
	if verbose != nil && *verbose {
		_, err = fmt.Fprintf(w, "%7s: ", vprefix)
		if err != nil {
			return err
		}
	}
	_, err = w.Write(text)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte{'\n'})
	if err != nil {
		return err
	}

	return nil
}

func fopt(ktype string) string {
	return fmt.Sprintf("%s-file", ktype)
}

func sopt(ktype string) string {
	return fmt.Sprintf("%s-suppress", ktype)
}

func getKeyoutSpec(ktype string) string {
	return fmt.Sprintf(
		"[--%s | --%s=<path>]",
		sopt(ktype), fopt(ktype),
	)
}

func getKeyoutClosure(cmd *cli.Cmd, ktype string) func(key signature.Key, verbose *bool) error {
	var (
		file = cmd.StringOpt(
			fopt(ktype), "",
			fmt.Sprintf("write the %s key to a new file at the specified path", ktype),
		)
		suppress = cmd.BoolOpt(
			sopt(ktype), false,
			fmt.Sprintf("suppress output of the %s key", ktype),
		)
	)

	return func(key signature.Key, verbose *bool) error {
		var w io.Writer = os.Stdout
		if suppress != nil && *suppress {
			w = ioutil.Discard
		}
		if file != nil && len(*file) > 0 {
			w, err := os.Create(*file)
			if err != nil {
				return err
			}
			defer w.Close()
		}

		return writeKey(w, key, verbose, ktype)
	}
}

func cmdEdNew(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"[-v] %s %s",
		getKeyoutSpec("private"),
		getKeyoutSpec("public"),
	)

	var (
		verbose = cmd.BoolOpt("v verbose", false, "verbose output")
		pvtOut  = getKeyoutClosure(cmd, "private")
		pubOut  = getKeyoutClosure(cmd, "public")
	)

	cmd.Action = func() {
		public, private, err := signature.Generate(signature.Ed25519, nil)
		check(err)
		pvterr := pvtOut(&private, verbose)
		puberr := pubOut(&public, verbose)
		check(pvterr)
		check(puberr)
	}
}

func cmdEdRaw(cmd *cli.Cmd) {
	cmd.Command("public", "transform a raw ed25519 public key into ndau format", cmdEdRawPublic)
	cmd.Command("signature", "transform a raw ed25519 signature into ndau format", cmdEdRawSig)
}

func cmdEdRawPublic(cmd *cli.Cmd) {
	cmd.Spec = getDataSpec(true)

	getData := getDataClosure(cmd, true)

	cmd.Action = func() {
		data := getData()

		key, err := signature.RawPublicKey(signature.Ed25519, data, nil)
		check(err)

		data, err = key.MarshalText()
		check(err)
		fmt.Println(string(data))
	}
}

func cmdEdRawSig(cmd *cli.Cmd) {
	cmd.Spec = getDataSpec(true)

	getData := getDataClosure(cmd, true)

	cmd.Action = func() {
		data := getData()

		sig, err := signature.RawSignature(signature.Ed25519, data)
		check(err)

		data, err = sig.MarshalText()
		check(err)
		fmt.Println(string(data))
	}
}
