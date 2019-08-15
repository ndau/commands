package main

import (
	"encoding/json"
	"fmt"

	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"

	"github.com/oneiro-ndev/commands/cmd/meic/ots/bitmart"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/sirupsen/logrus"
)

// this list is broken out into a separate file to make it easy to find when
// adding a new OTS implementation.

// this list contains an instance of each implemented order tracking system.
// the IUS refers to it when initializing in order to launch all instances

type Exchange struct {
	Symbol     string
	APIKeyPath string
}

func updateQty(order SellOrder) {
	fmt.Println("update = ", order)
}
func delete(order SellOrder) {
	fmt.Println("delete = ", order)
}
func submit(order SellOrder) {
	fmt.Println("submit = ", order)
}

func (e Exchange) Run(logger logrus.FieldLogger, sales chan<- TargetPriceSale, updates <-chan UpdateOrders) {
	fmt.Println("symbol = ", e.Symbol)
	upd := <-updates
	for idx := range upd.Orders {
		fmt.Println("order = ", upd.Orders[idx])
	}
	key, err := bitmart.LoadAPIKey(e.APIKeyPath)
	check(err, "loading api key")
	auth := bitmart.NewAuth(key)

	statusFilter := bitmart.OrderStatusFrom("pendingandpartialsuccess")
	fmt.Println("using order status filter:", statusFilter)
	orders, err := bitmart.GetOrderHistory(&auth, e.Symbol, statusFilter)
	check(err, "getting orders")

	data, err := json.MarshalIndent(orders, "", "  ")
	check(err, "formatting output")

	curStack := make([]SellOrder, 0, 4)

	for i := len(orders) - 1; i >= 0; i-- {
		if orders[i].Side == "sell" {
			fQty := fmt.Sprintf("%f", orders[i].RemainingAmount)
			qty, err := math.ParseNdau(fQty)
			check(err, "converting remaining amount")

			fPrice := fmt.Sprintf("%f", orders[i].Price)
			price, err := pricecurve.ParseDollars(fPrice)
			check(err, "converting price")

			curStack = append(curStack, SellOrder{
				Qty:   qty,
				Price: price,
			})
		}
	}

	SynchronizeOrders(curStack, upd.Orders, updateQty, delete, submit)

	fmt.Println(string(data))
}

var otsImpls = []OrderTrackingSystem{Exchange{"XND_USDT", "bitmart.apikey.json"}}
