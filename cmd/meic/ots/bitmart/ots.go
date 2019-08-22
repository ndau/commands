package bitmart

import (
	"fmt"
	"time"

	"github.com/oneiro-ndev/commands/cmd/meic/ots"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// An OTS is the bitmart implementation of the OTS interface
type OTS struct {
	Symbol       string
	args         BMArgs
	auth         Auth
	statusFilter OrderStatus
}

// compile-time check that we actually do implement that interface
var _ ots.OrderTrackingSystem = (*OTS)(nil)

// UpdateQty updates a bitmart order, setting a new qty on offer at the same price
func (e OTS) UpdateQty(order ots.SellOrder) error {
	err := CancelOrder(&e.auth, order.ID)
	if err != nil {
		return errors.Wrap(err, "cancel order request")
	}
	qty := float64(order.Qty) / 100000000
	price := float64(order.Price) / 100000000000
	id, err := PlaceOrder(&e.auth, e.Symbol, "sell", price, qty)
	order.ID = uint64(id)
	return err
}

// Delete deletes a bitmart sell order
func (e OTS) Delete(order ots.SellOrder) error {
	return CancelOrder(&e.auth, order.ID)
}

// Submit submits a new bitmart sell order
func (e OTS) Submit(order ots.SellOrder) error {
	qty := float64(order.Qty) / 100000000
	price := float64(order.Price) / 100000000000
	id, err := PlaceOrder(&e.auth, e.Symbol, "sell", price, qty)
	order.ID = uint64(id)
	return err
}

// Init implements ots.OrderTrackingSystem
func (e *OTS) Init(logger logrus.FieldLogger, args interface{}) error {
	logger = logger.WithField("ots", "bitmart")

	if bma, ok := args.(HasBMArgs); ok {
		e.args = bma.GetBMArgs()
	} else {
		return errors.New("args did not implement HasBMArgs")
	}

	key, err := LoadAPIKey(e.args.APIKeyPath)
	if err != nil {
		return errors.Wrap(err, "bitmart ots: loading api key")
	}
	e.auth = NewAuth(key)

	e.statusFilter = OrderStatusFrom("pendingandpartialsuccess")
	logger.WithFields(logrus.Fields{
		"statusFilter": e.statusFilter,
	}).Debug("setup status filter")

	return nil
}

// Run implements ots.OrderTrackingSystem
func (e *OTS) Run(
	logger logrus.FieldLogger,
	sales chan<- ots.TargetPriceSale,
	updates <-chan ots.UpdateOrders,
	errs chan<- error,
) {
	logger = logger.WithField("ots", "bitmart")

	// launch a goroutine to watch the updates channel
	go func() {
		logger = logger.WithField("goroutine", "updates monitor")
		for {
			// notice any update instructions
			upd := <-updates

			logger.WithField("desired stack", upd.Orders).Debug("received update instruction")

			// bitmart truncates the precision of Qty and Price, so we have to reduce
			// both of those in order for synchronization to work. However, we can't
			// modify upd, as that is a shared data structure which will affect other
			// exchanges. We therefore copy the data, truncating en-route.
			desiredStack := make([]ots.SellOrder, len(upd.Orders))
			for idx, o := range upd.Orders {
				desiredStack[idx].Qty = o.Qty - (o.Qty % 1000000)
				desiredStack[idx].Price = o.Price - (o.Price % 10000000)
			}

			// update the current stack
			orders, err := GetOrderHistory(&e.auth, e.Symbol, e.statusFilter)
			if err != nil {
				errs <- errors.Wrap(err, "getting orders")
				return
			}

			curStack := make([]ots.SellOrder, 0, len(orders))

			for _, order := range orders {
				if !order.IsSale() {
					continue
				}

				fQty := fmt.Sprintf("%f", order.RemainingAmount)
				qty, err := math.ParseNdau(fQty)
				if err != nil {
					errs <- errors.Wrap(err, "converting remaining amount")
					return
				}

				fPrice := fmt.Sprintf("%f", order.Price)
				price, err := pricecurve.ParseDollars(fPrice)
				if err != nil {
					errs <- errors.Wrap(err, "converting price")
					return
				}

				curStack = append(curStack, ots.SellOrder{
					Qty:   qty,
					Price: price,
					ID:    uint64(order.EntrustID),
				})
			}

			// we've defined our order manipulation functions as methods
			// on this struct, but the signature isn't quite the same.
			// this wrapper converts the signature appropriately.
			ewrap := func(f func(ots.SellOrder) error) func(ots.SellOrder) {
				return func(s ots.SellOrder) {
					err := f(s)
					if err != nil {
						errs <- err
					}
				}
			}

			ots.SynchronizeOrders(
				curStack,
				desiredStack,
				ewrap(e.UpdateQty),
				ewrap(e.Delete),
				ewrap(e.Submit),
			)
		}
	}()

	logger = logger.WithField("goroutine", "trade poller")

	var err error

	// make first call to get max trade ID
	var maxTradeID int64
	_, maxTradeID, err = GetTradeHistory(&e.auth, e.Symbol)
	if err != nil {
		errs <- errors.Wrap(err, "get order history")
		return
	}
	var trades []Trade
	for {
		logger.WithField("max trade id", maxTradeID).Debug("polling for new trades")
		trades, maxTradeID, err = GetTradeHistoryAfter(&e.auth, e.Symbol, maxTradeID)
		if err != nil {
			errs <- errors.Wrap(err, "get order history after")
			return
		}
		for _, trade := range trades {
			var tps = ots.TargetPriceSale{Qty: trade.Amount}
			sales <- tps
		}
		time.Sleep(5 * time.Second)
	}

}
