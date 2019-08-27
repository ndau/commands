package main

import (
	"fmt"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/recovery/pkg/signer"
	"github.com/sirupsen/logrus"
)

type args struct {
	ServerAddr        string               `arg:"-s,--server-addr,required"`
	ServerPubKey      signature.PublicKey  `arg:"-p,--server-pub-key,required"`
	ServerPvtKey      signature.PrivateKey `arg:"-P,--server-pvt-key,required"`
	NodeAddr          string               `arg:"-n,--node-addr,required"`
	DefaultStackDepth uint                 `arg:"-d,--default-stack-depth"`
	ConfPath          string               `arg:"-c,--config-path"`
}

func main() {
	args := args{
		DefaultStackDepth: 3,
	}
	arg.MustParse(&args)

	logger = logrus.New().WithField("bin", "meic")
	fmt.Println("pubkey = ", args.ServerPubKey.FullString())
	fmt.Println("pvtkey = ", args.ServerPvtKey.FullString())
	device, err := signer.NewVirtualDevice(args.ServerPubKey.FullString(), args.ServerPvtKey.FullString())
	check(err, "creating virtual signer device")

	config, err := LoadConfig(args.ConfPath)
	check(err, "loading configuration")

	ius, err := NewIUS(
		logger.(*logrus.Entry),
		args.ServerAddr,
		device,
		args.NodeAddr,
		args.DefaultStackDepth,
		config,
	)
	check(err, "creating ius")

	err = ius.Run(nil)
	check(err, "running ius")
}
