package main

import (
	"os"

	"github.com/oneiro-ndev/chaos/pkg/chaos/config"
)

func setNdaunodeF(s *string) {
	if s != nil && len(*s) > 0 {
		cp := config.DefaultConfigPath(getNdauhome())
		conf, err := config.LoadDefault(cp)
		check(err)
		conf.NdauAddress = *s
		check(conf.Dump(cp))
		os.Exit(0)
	}
}

func unsetNdaunodeF(b *bool) {
	if b != nil && *b {
		cp := config.DefaultConfigPath(getNdauhome())
		conf, err := config.LoadDefault(cp)
		check(err)
		conf.NdauAddress = ""
		check(conf.Dump(cp))
		os.Exit(0)
	}
}
