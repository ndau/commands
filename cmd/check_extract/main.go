package main

import (
	"fmt"
	"os"

	util "github.com/oneiro-ndev/genesis/pkg/cli.util"
	"github.com/oneiro-ndev/genesis/pkg/config"
	"github.com/oneiro-ndev/genesis/pkg/etl"
)

func main() {
	path := config.DefaultConfigPath(util.GetNdauhome())
	var rows []etl.RawRow
	var err error
	err = config.WithConfig(path, func(conf *config.Config) error {
		rows, err = etl.Extract(conf)
		if err != nil {
			return err
		}
		return nil
	})
	util.Check(err)
	fmt.Println("Rows extracted:", len(rows))
	fmt.Printf("First row:  %v\n", rows[0])
	if len(rows) > 0 {
		fmt.Printf("Middle row: %v\n", rows[len(rows)/2])
		fmt.Printf("Last row:   %v\n", rows[len(rows)-1])
	}

	duplicates := etl.DuplicateAddresses(rows)
	if len(duplicates) > 0 {
		fmt.Println("WARN: duplicate addresses present:")
		for addr, rows := range duplicates {
			fmt.Printf("  %s:\n", addr)
			fmt.Printf("    ")
			for _, row := range rows {
				fmt.Printf("%d ", row)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("All addresses are distinct")
	}

	foundNERow := false
	for _, row := range rows {
		if row.RewardTarget != nil {
			foundNERow = true
			fmt.Printf("First row with non-empty target: %s\n", row)
			break
		}
	}
	if !foundNERow {
		fmt.Println("No rows found with non-empty reward target")
	}

	// verify that all rows either send or receive rows but not both
	inbounds := make(map[string][]string)
	for _, row := range rows {
		if row.RewardTarget != nil {
			inbounds[*row.RewardTarget] = append(
				inbounds[*row.RewardTarget],
				row.Address,
			)
		}
	}

	bidirectionalTransfers := 0
	for _, row := range rows {
		if row.RewardTarget != nil && inbounds[row.Address] != nil {
			bidirectionalTransfers++
			fmt.Printf(
				"ERR: Row %d, address %s: transfers to %s and receives inbounds from the following:\n",
				row.RowNumber,
				row.Address,
				*row.RewardTarget,
			)
			for _, inbound := range inbounds[*row.RewardTarget] {
				fmt.Println("\t", inbound)
			}
		}
	}
	if bidirectionalTransfers > 0 {
		os.Exit(1)
	}
}
