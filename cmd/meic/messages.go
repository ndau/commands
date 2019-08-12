package main

import (
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// A TargetPriceSale message is transmitted from an OTS instance to the IUS
// whenever a target price sale occurs.
type TargetPriceSale struct {
	Qty math.Ndau
}

// A SellOrder represents a qty of ndau to offer for sale at a particular price.
//
// Its ID is never set or read by the IUS; it is instead intended to be used by OTS
// instances for disambiguation when synchronizing the desired and current orders.
type SellOrder struct {
	Qty   math.Ndau
	Price pricecurve.Nanocent
	ID    uint64
}

// An UpdateOrders message is transmitted from the IUS to each OTS instance
// whenever the OTS's offers should be updated.
type UpdateOrders struct {
	Orders []SellOrder
}
