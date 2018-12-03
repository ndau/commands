package main

import (
	"fmt"

	util "github.com/oneiro-ndev/genesis/pkg/cli.util"
	"github.com/oneiro-ndev/genesis/pkg/config"
)

func main() {
	path := config.DefaultConfigPath(util.GetNdauhome())
	err := config.WithConfig(path, func(c *config.Config) error {
		return c.CheckColumns()
	})
	util.Check(err)
	fmt.Println(path)
}
