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
	// for the moment, we'll hardcode a stack of 3 full blocks for sale past the
	// current one. We can figure out a good way to make this configurable later.
	stack := make([]SellOrder, 0, 4)
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
		stack = append(stack, SellOrder{
			Qty:   remainingInBlock,
			Price: price(issued),
		})
		issued += remainingInBlock
	}
	for i := 0; i < 2; i++ {
		stack = append(stack, SellOrder{
			Qty:   napuInBlock,
			Price: price(issued),
		})
		issued += napuInBlock
	}

	// 3. send that stack individually to each OTS
	uos := UpdateOrders{
		Orders: stack,
	}

	for _, uoChan := range ius.updates {
		// spawn goroutines because we don't want to block the main thread
		// in case any of the OTSs are blocked
		go func(c chan<- UpdateOrders) {
			c <- uos
		}(uoChan)
	}
}
