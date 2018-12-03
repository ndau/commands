package main

import (
	"os"

	cli "github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("keytool", "manipulate key strings on the command line")

	app.Command("hd secp256k1", "manipulate HD keys", hd)
	app.Command("ed ed25519", "manipulate ed25519 keys", ed)
	app.Command("sign", "sign some data", cmdSign)
	app.Command("verify", "verify some data", cmdVerify)
	app.Command("addr", "generate addresses from public keys", cmdAddr)
	app.Command("truncate", "remove any extra data from a key", cmdTruncate)
	app.Command("inspect", "inspect a key", cmdInspect)

	app.Run(os.Args)
}

// hd subcommand
func hd(cmd *cli.Cmd) {
	cmd.Command("new", "create a new HD key", cmdHDNew)
	cmd.Command("public", "create a public key from supplied key", cmdHDPublic)
	cmd.Command("child", "create a child key derived from the supplied key", cmdHDChild)
	cmd.Command("convert", "convert an old-format key into the new format", cmdHDConvert)
	cmd.Command("addr", "convert HD key to address", cmdHDAddr)
}

// ed subcommand
func ed(cmd *cli.Cmd) {
	cmd.Command("new", "create a new ed25519 keypair", cmdEdNew)
	cmd.Command("raw", "create an ndau-style ed25519 public key from raw bytes", cmdEdRaw)
}
