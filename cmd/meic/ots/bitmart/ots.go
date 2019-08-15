package bitmart

import (
	"fmt"

	"github.com/oneiro-ndev/commands/cmd/meic/ots"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// An OTS is the bitmart implementation of the OTS interface
type OTS struct {
	Symbol       string
	APIKeyPath   string
	auth         Auth
	statusFilter OrderStatus
}

// compile-time check that we actually do implement that interface
var _ ots.OrderTrackingSystem = (*OTS)(nil)

func updateQty(order ots.SellOrder) {
	fmt.Println("update = ", order)
}
func delete(order ots.SellOrder) {
	fmt.Println("delete = ", order)
}
func submit(order ots.SellOrder) {
	fmt.Println("submit = ", order)
}

// Init implements ots.OrderTrackingSystem
func (e OTS) Init(logger logrus.FieldLogger) error {
	fmt.Println("symbol = ", e.Symbol)

	key, err := LoadAPIKey(e.APIKeyPath)
	if err != nil {
		return errors.Wrap(err, "bitmart ots: loading api key")
	}
	e.auth = NewAuth(key)

	e.statusFilter = OrderStatusFrom("pendingandpartialsuccess")
	logger.WithFields(logrus.Fields{
		"ots":          "bitmart",
		"statusFilter": e.statusFilter,
	}).Debug("setup status filter")

	return nil
}

// Run implements ots.OrderTrackingSystem
func (e OTS) Run(
	logger logrus.FieldLogger,
	sales chan<- ots.TargetPriceSale,
	updates <-chan ots.UpdateOrders,
	errs chan<- error,
) {
	logger = logger.WithField("ots", "bitmart")

	// launch a goroutine to watch the updates channel
	go func() {
		logger = logger.WithField("goroutine", "OTS updates monitor")
		for {
			// notice any update instructions
			upd := <-updates

			logger.WithField("desired stack", upd.Orders).Debug("received update instruction")

			// update the current stack
			orders, err := GetOrderHistory(&e.auth, e.Symbol, e.statusFilter)
			if err != nil {
				errs <- errors.Wrap(err, "getting orders")
				return
			}

			curStack := make([]ots.SellOrder, 0, 4)

			for i := len(orders) - 1; i >= 0; i-- {
				if orders[i].Side == "sell" {
					fQty := fmt.Sprintf("%f", orders[i].RemainingAmount)
					qty, err := math.ParseNdau(fQty)
					if err != nil {
						errs <- errors.Wrap(err, "converting remaining amount")
						return
					}

					fPrice := fmt.Sprintf("%f", orders[i].Price)
					price, err := pricecurve.ParseDollars(fPrice)
					if err != nil {
						errs <- errors.Wrap(err, "converting price")
						return
					}

					curStack = append(curStack, ots.SellOrder{
						Qty:   qty,
						Price: price,
					})
				}
			}

			ots.SynchronizeOrders(curStack, upd.Orders, updateQty, delete, submit)

		}
	}()

	// TODO: set up and monitor a websocket connection to bitmart, sending on
	// the sales channel as appropriate
}
