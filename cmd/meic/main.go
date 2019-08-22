package main

import (
	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/recovery/pkg/signer"
	"github.com/sirupsen/logrus"
)

func main() {
	args := args{
		DefaultStackDepth: 3,
	}
	arg.MustParse(&args)

	logger = logrus.New().WithField("bin", "meic")
	device, err := signer.NewVirtualDevice(args.ServerPubKey.FullString(), args.ServerPvtKey.FullString())
	check(err, "creating virtual signer device")

	config, err := LoadConfig(args.ConfPath)
	check(err, "loading configuration")

	ius, err := NewIUS(
		logger.(*logrus.Entry),
		args.ServerAddr,
		device,
		args,
		config,
	)
	check(err, "creating ius")

	err = ius.Run(nil)
	check(err, "running ius")
}
