package main

import (
	"os"

	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
)

func setChaosnodeF(s *string) {
	if s != nil && len(*s) > 0 {
		cp := config.DefaultConfigPath(getNdauhome())
		conf, err := config.LoadDefault(cp)
		check(err)
		conf.ChaosAddress = *s
		check(conf.Dump(cp))
		os.Exit(0)
	}
}

func unsetChaosnodeF(b *bool) {
	if b != nil && *b {
		cp := config.DefaultConfigPath(getNdauhome())
		conf, err := config.LoadDefault(cp)
		check(err)
		conf.ChaosAddress = ""
		check(conf.Dump(cp))
		os.Exit(0)
	}
}
