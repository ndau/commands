package main

import (
	"fmt"
	"math"

	"github.com/oneiro-ndev/commands/cmd/meic/ots"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	ndaumath "github.com/oneiro-ndev/ndaumath/pkg/types"
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
	fmt.Println("summary = ", summary)
	// 2. compute the current desired target sales stack
	stack := make([]ots.SellOrder, 0, ius.stackGen+1)
	partial := uint(0)
	// JSG round remaining to 2 significant digs of ndau for exchange
	issued := ndaumath.Ndau(math.Round(float64(summary.TotalIssue)/1000000) * 1000000)

	napuInBlock := ndaumath.Ndau(pricecurve.SaleBlockQty * constants.NapuPerNdau)
	issuedInBlock := issued % napuInBlock
	remainingInBlock := (napuInBlock - issuedInBlock) % napuInBlock

	price := func(issued ndaumath.Ndau) pricecurve.Nanocent {
		p, err := pricecurve.PriceAtUnit(issued)
		if err != nil {
			check(err, "calculating expected price")
		}
		return p
	}

	if remainingInBlock > 0 {
		partial = 1
		stack = append(stack, ots.SellOrder{
			Qty:   remainingInBlock,
			Price: price(issued),
		})
		issued += remainingInBlock
	}
	for i := partial; i < ius.stackGen; i++ {
		stack = append(stack, ots.SellOrder{
			Qty:   napuInBlock,
			Price: price(issued),
		})
		issued += napuInBlock
	}

	// 3. send that stack individually to each OTS
	for idx, uoChan := range ius.updates {
		depth := ius.stackDefault
		if ius.config != nil && ius.config.DepthOverrides != nil {
			if do, ok := ius.config.DepthOverrides[uint(idx)]; ok {
				depth = do
			}
		}
		// spawn goroutines because we don't want to block the main thread
		// in case any of the OTSs are blocked
		go func(c chan<- ots.UpdateOrders, depth int) {
			c <- ots.UpdateOrders{
				Orders: stack[:depth],
			}
		}(uoChan, int(depth))
	}
}
