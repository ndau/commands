package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto/ed25519"
	amino "github.com/tendermint/tendermint/crypto/encoding/amino"
)

func getCVC(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "(PUBKEY | -x=<PUBKEY_HEX>) POWER"

		pk64 := cmd.StringArg("PUBKEY", "", "padding-free base64 encoding of Tendermint ed25519 public key")
		pkx := cmd.StringOpt("x hex", "", "hexadecimal encoding of Tendermint ed25519 public key")
		power := cmd.IntArg("POWER", 0, "power to assign to this node")

		cmd.Action = func() {
			var pkb []byte
			var err error

			switch {
			case pkx != nil && len(*pkx) > 0:
				pkb, err = hex.DecodeString(*pkx)
			case pk64 != nil && len(*pk64) > 0:
				pkb, err = base64.RawStdEncoding.DecodeString(*pk64)
			default:
				err = errors.New("PUBKEY must be set")
			}
			orQuit(err)

			if len(pkb) != ed25519.PubKeyEd25519Size {
				// if we got a straight ed25519 key, just pass it through
				// chances are good that we got something amino-encoded,
				// though, so let's try parsing that

				pk, err := amino.PubKeyFromBytes(pkb)
				orQuit(err)

				pke, ised25519 := pk.(ed25519.PubKeyEd25519)
				if !ised25519 {
					orQuit(errors.New("PUBKEY must be of type Ed25519"))
				}
				pkb = []byte(pke[:])
			}

			if *power < 0 {
				orQuit(errors.New("cvc POWER must be > 0"))
			}

			if *verbose {
				fmt.Printf("CommandValidatorChange: PubKey %x (%d bytes) Power %d\n", pkb, len(pkb), *power)
			}

			conf := getConfig()
			if conf.CVC == nil {
				orQuit(errors.New("CVC data not set in tool config"))
			}

			fkeys := config.FilterK(conf.CVC.Keys, keys)

			cvc := ndau.NewCommandValidatorChange(
				pkb, int64(*power),
				sequence(conf, conf.CVC.Address),
				fkeys,
			)

			result, err := tool.SendCommit(tmnode(conf.Node), &cvc)
			finish(*verbose, result, err, "cvc")
		}
	}
}
