package main

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
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

func cmdEdNode(cmd *cli.Cmd) {
	cmd.Command("new", "generate new keys and create an ndau node data file", cmdEdNodeNew)
	cmd.Command("private", "use an ed25519 private key to generate an ndau node data file", cmdEdNodePrivate)
}

type tarFile struct {
	Name string
	Body *bytes.Buffer
}

func createNodeKeyFile(priv signature.PrivateKey) tarFile {
	templ := `{"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"%s"}}`
	b := base64.StdEncoding.EncodeToString(priv.KeyBytes())
	return tarFile{
		Name: "node_key.json",
		Body: bytes.NewBufferString(fmt.Sprintf(templ, b)),
	}
}

func createPrivKeyFile(public signature.PublicKey, private signature.PrivateKey) tarFile {
	templ := `
{
  "address": "%X",
  "pub_key": {
    "type": "tendermint/PubKeyEd25519",
    "value": "%s"
  },
  "priv_key": {
    "type": "tendermint/PrivKeyEd25519",
    "value": "%s"
  }
}
`
	pub := base64.StdEncoding.EncodeToString(public.KeyBytes())
	prv := base64.StdEncoding.EncodeToString(private.KeyBytes())
	sum := sha256.Sum256(public.KeyBytes())
	addr := sum[:20]
	return tarFile{
		Name: "priv_validator_key.json",
		Body: bytes.NewBufferString(fmt.Sprintf(templ, addr, pub, prv)),
	}
}

func createNdauAccountFile(public signature.PublicKey, private signature.PrivateKey, kind byte) tarFile {
	templ := `
{
  "address": "%s",
  "ownership_public_key": "%s",
  "ownership_private_key": "%s"
}
`

	pub, err := public.MarshalText()
	check(err)
	prv, err := private.MarshalText()
	check(err)

	addr, err := address.Generate(kind, public.KeyBytes())
	check(err)

	return tarFile{
		Name: "ndau_info.json",
		Body: bytes.NewBufferString(fmt.Sprintf(templ, addr, pub, prv)),
	}
}

func writeTGZ(filename string, tfs ...tarFile) error {
	// Create and add some files to the archive.
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	tw := tar.NewWriter(f)
	for _, file := range tfs {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(file.Body.Len()),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write(file.Body.Bytes()); err != nil {
			return err
		}
	}
	return tw.Close()
}

func cmdEdNodeNew(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"%s [--filename]",
		getKindSpec(),
	)

	getKind := getKindClosure(cmd)

	var (
		fname = cmd.StringOpt("filename", "node-identity.tgz", "write the output as a .tgz file to this name")
	)

	cmd.Action = func() {
		public1, private1, err := signature.Generate(signature.Ed25519, nil)
		check(err)
		public2, private2, err := signature.Generate(signature.Ed25519, nil)
		check(err)
		kind := getKind()

		tf1 := createNodeKeyFile(private1)
		tf2 := createPrivKeyFile(public2, private2)
		tf3 := createNdauAccountFile(public1, private1, kind)

		err = writeTGZ(*fname, tf1, tf2, tf3)
		check(err)
	}
}

// match indicates whether or not given public and private keys match each other.
// This should be replaced with signature.match when we can safely depend on an ndautool >1.2.1
func match(pub signature.PublicKey, pvt signature.PrivateKey) bool {
	// it's kind of silly, but the best way we have to tell if the keys match
	// is just to sign some data, and then attempt to verify it
	data := make([]byte, 64)
	_, err := rand.Read(data)
	if err != nil {
		return false
	}
	sig := pvt.Sign(data)
	return pub.Verify(data, sig)
}

func cmdEdNodePrivate(cmd *cli.Cmd) {
	cmd.Spec = fmt.Sprintf(
		"%s %s [--filename]",
		getKeySpec("PVT"),
		getKindSpec(),
	)

	var (
		fname = cmd.StringOpt("filename", "node-identity.tgz", "write the output as a .tgz file to this name")
	)

	getKey := getKeyClosure(cmd, "PVT", "private ownership key for the ndau account associated with the node")
	getKind := getKindClosure(cmd)

	cmd.Action = func() {
		// I'm not convinced this is the right way to go about things yet.
		k := getKey()
		private1, err := signature.RawPrivateKey(signature.Ed25519, k.KeyBytes(), k.ExtraBytes())
		fmt.Println(len(k.KeyBytes()), len(k.ExtraBytes()), len(private1.KeyBytes()))
		pubBytes := private1.KeyBytes()[32:]
		public1, err := signature.RawPublicKey(signature.Ed25519, pubBytes, nil)
		check(err)
		if !match(*public1, *private1) {
			fmt.Println("keys don't match!")
			os.Exit(1)
		}

		public2, private2, err := signature.Generate(signature.Ed25519, nil)
		check(err)
		kind := getKind()

		tf1 := createNodeKeyFile(*private1)
		tf2 := createPrivKeyFile(public2, private2)
		tf3 := createNdauAccountFile(*public1, *private1, kind)

		err = writeTGZ(*fname, tf1, tf2, tf3)
		check(err)
	}
}
