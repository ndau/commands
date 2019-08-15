package main

import (
	"fmt"

	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// this helper computes the desired stack and forwards it to all OTSs
func (ius *IssuanceUpdateSystem) updateOTSs() {
	// 1. get the total issuance from the blockchain
	summary, _, err := tool.GetSummary(ius.tmNode)
	if err != nil {
		// maybe we should figure out a better error-handling solution than all
		// these panics
		check(err, "failed to get blockchain summary")
	}

	// 2. compute the current desired target sales stack
	stack := make([]SellOrder, 0, ius.stackGen+1)
	partial := uint(0)
	issued := summary.TotalIssue
	fmt.Println("issued =", issued)

	napuInBlock := math.Ndau(pricecurve.SaleBlockQty * constants.NapuPerNdau)
	issuedInBlock := issued % napuInBlock
	fmt.Println("issued in block =", issuedInBlock)
	remainingInBlock := (napuInBlock - issuedInBlock) % napuInBlock
	fmt.Println("remaining in block =", remainingInBlock)

	price := func(issued math.Ndau) pricecurve.Nanocent {
		p, err := pricecurve.PriceAtUnit(issued)
		if err != nil {
			check(err, "calculating expected price")
		}
		return p
	}

	if remainingInBlock > 0 {
		partial = 1
		stack = append(stack, SellOrder{
			Qty:   remainingInBlock,
			Price: price(issued),
		})
		issued += remainingInBlock
	}
	for i := uint(0); i < ius.stackGen; i++ {
		stack = append(stack, SellOrder{
			Qty:   napuInBlock,
			Price: price(issued),
		})
		issued += napuInBlock
	}

	// 3. send that stack individually to each OTS
	for idx, uoChan := range ius.updates {
		depth := ius.stackDefault + partial
		if ius.config != nil && ius.config.DepthOverrides != nil {
			if do, ok := ius.config.DepthOverrides[uint(idx)]; ok {
				depth = do + partial
			}
		}
		// spawn goroutines because we don't want to block the main thread
		// in case any of the OTSs are blocked
		go func(c chan<- UpdateOrders, depth int) {
			c <- UpdateOrders{
				Orders: stack[:depth],
			}
		}(uoChan, int(depth))
	}
}
