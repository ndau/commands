package main

import (
	"github.com/oneiro-ndev/commands/cmd/meic/ots"
	"github.com/oneiro-ndev/commands/cmd/meic/ots/bitmart"
)

// this list is broken out into a separate file to make it easy to find when
// adding a new OTS implementation.

// this list contains an instance of each implemented order tracking system.
// the IUS refers to it when initializing in order to launch all instances
var otsImpls = []ots.OrderTrackingSystem{
	bitmart.OTS{
		Symbol:     "NDAU_USDT",
		APIKeyPath: "test.apikey.json",
	},
}
