package main

import (
	"errors"
	"fmt"

	util "github.com/oneiro-ndev/genesis/pkg/cli.util"
	"github.com/oneiro-ndev/genesis/pkg/config"
	"github.com/oneiro-ndev/genesis/pkg/etl"
)

func main() {
	ndauhome := util.GetNdauhome()
	path := config.DefaultConfigPath(ndauhome)
	var rows []etl.RawRow
	var err error
	err = config.WithConfig(path, func(conf *config.Config) error {
		rows, err = etl.Extract(conf)
		if err != nil {
			return err
		}
		duplicates := etl.DuplicateAddresses(rows)
		if len(duplicates) > 0 {
			fmt.Println("ERROR: duplicate addresses present:")
			for addr, rows := range duplicates {
				fmt.Printf("  %s:\n", addr)
				fmt.Printf("    ")
				for _, row := range rows {
					fmt.Printf("%d ", row)
				}
				fmt.Println()
			}
			return errors.New("duplicate addresses")
		}

		return etl.Load(conf, rows, ndauhome)
	})
	util.Check(err)
}
