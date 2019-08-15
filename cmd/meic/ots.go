package main

import (
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
)

// An OrderTrackingSystem handles all the details associated with an individual
// exchange.
//
// It must make a best-effort attempt to generate a TargetPriceSale
// message as close to real-time as possible after a target price sale, and
// must respond to UpdateOrders messages by adjusting that exchange's open
// sell orders according to the message.
type OrderTrackingSystem interface {
	// Run is used to start an OTS instance.
	//
	// The sales channel has a small buffer, but in the event the buffer fills,
	// OTS instances must block until it can add the sale to the channel.
	// Otherwise, a sale could fall through the cracks and never generate
	// an appropriate issuance.
	Run(logger logrus.FieldLogger, sales chan<- TargetPriceSale, updates <-chan UpdateOrders)
}

// SynchronizeOrders handles the grunt work of diffing out the updates implied
// by a current and desired set of sell orders.
func SynchronizeOrders(
	current, desired []SellOrder,
	updateQty func(SellOrder),
	delete func(SellOrder),
	submit func(SellOrder),
) {
	// sort the current and desired slices by price
	sort.Slice(current, func(i, j int) bool { return current[i].Price < current[j].Price })
	sort.Slice(desired, func(i, j int) bool { return desired[i].Price < desired[j].Price })

	fmt.Println("current =", current)
	fmt.Println("desired =", desired)
	// in essence, this is a merge sort on current and desired
	ci := 0
	di := 0
	for ci < len(current) && di < len(desired) {
		switch {
		case current[ci].Price < desired[di].Price:
			// there is ndau for sale at too low a price
			delete(current[ci])
			ci++
		case current[ci].Price == desired[di].Price:
			// the price is right
			if current[ci].Qty != desired[di].Qty {
				current[ci].Qty = desired[di].Qty
				updateQty(current[ci])
			}
			ci++
			di++
		case current[ci].Price > desired[di].Price:
			// there is ndau for sale at too high a price, probably
			// check it on the next iteration
			di++
		}
	}
	// now remove any extra current orders which haven't been dealt with
	for ; ci < len(current); ci++ {
		delete(current[ci])
	}
	// now add any extra desired orders which haven't been dealt with
	for ; di < len(desired); di++ {
		submit(desired[di])
	}
}
