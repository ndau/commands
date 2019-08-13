package main

import (
	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/recovery/pkg/signer"
	"github.com/sirupsen/logrus"
)

type args struct {
	ServerAddr   string               `arg:"-s,--server-addr,required"`
	ServerPubKey signature.PublicKey  `arg:"-p,--server-pub-key,required"`
	ServerPvtKey signature.PrivateKey `arg:"-P,--server-pvt-key,required"`
	NodeAddr     string               `arg:"-n,--node-addr,required"`
}

func main() {
	var args args
	arg.MustParse(&args)

	logger = logrus.New().WithField("bin", "meic")
	device, err := signer.NewVirtualDevice(args.ServerPubKey.FullString(), args.ServerPvtKey.FullString())
	check(err, "creating virtual signer device")

	ius, err := NewIUS(logger.(*logrus.Entry), args.ServerAddr, device, args.NodeAddr)
	check(err, "creating ius")

	err = ius.Run(nil)
	check(err, "running ius")
}
