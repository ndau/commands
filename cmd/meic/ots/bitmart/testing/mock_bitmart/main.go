package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	bitmart "github.com/oneiro-ndev/commands/cmd/meic/ots/bitmart"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

const port = 25204 // big-endian numeric interpretation of bytes of "bt"

func check(err error, context string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, context+":")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// chance is in the range [0,1]
func randomSide(chanceSell float64) string {
	if rand.Float64() < chanceSell {
		return bitmart.SideSell
	}
	return bitmart.SideBuy
}

// not completely random: we only care about sell orders on the ndau symbol,
// so several fields are not entirely random, or have non-uniform probabilities
func randomOrder(entrustID int64) bitmart.Order {
	amount := 1500 * rand.Float64()
	return bitmart.Order{
		EntrustID:       entrustID,
		Symbol:          bitmart.NdauSymbol,
		Timestamp:       bitmart.Time(),
		Side:            randomSide(0.85),
		Price:           17 + (3 * rand.Float64()),
		OriginalAmount:  amount,
		RemainingAmount: amount,
		Status:          int64(bitmart.Pending),
	}
}

// also not completely random. nil if no potential orders.
func randomTrade(orders map[int64]bitmart.Order, tradeID int64) *bitmart.Trade {
	potentialOrders := make([]bitmart.Order, 0, len(orders))
	for _, order := range orders {
		if order.RemainingAmount > 0 {
			potentialOrders = append(potentialOrders, order)
		}
	}
	if len(potentialOrders) == 0 {
		return nil
	}
	order := potentialOrders[rand.Intn(len(potentialOrders))]

	// 50% chance of taking the full portion, otherwise, some random portion
	portion := 2 * rand.Float64()
	if portion > 1 {
		portion = 1
	}

	qty := portion * order.RemainingAmount
	order.RemainingAmount -= qty
	order.ExecutedAmount += qty

	if order.RemainingAmount > 0 {
		order.Status = int64(bitmart.PartialSuccess)
	} else {
		order.Status = int64(bitmart.Success)
	}

	orders[order.EntrustID] = order

	return &bitmart.Trade{
		Symbol:    order.Symbol,
		Amount:    math.Ndau(qty * constants.NapuPerNdau),
		TradeID:   tradeID,
		Price:     qty * order.Price,
		Active:    true,
		EntrustID: order.EntrustID,
		Timestamp: bitmart.Time(),
	}
}

func main() {
	// make this deterministic, if possible
	rand.Seed(port)

	// make a bunch of orders
	orders := make(map[int64]bitmart.Order)
	for i := int64(1); i <= 100; i++ {
		orders[i] = randomOrder(i)
	}
	// make a bunch of trades
	trades := make(map[int64]bitmart.Trade)
	for i := int64(1); i <= 100; i++ {
		t := randomTrade(orders, i)
		if t == nil {
			break
		}
		trades[i] = *t
	}
	// we don't need to worry about ensuring that there are some unfilfilled
	// orders, because the issuance service never actually queries the orders list
	// endpoint

	// we're going to use a muxer with a nice shorthand for path parameters
	m := mux.NewRouter()

	// now define some handlers
	m.HandleFunc("/v2/authentication", func(w http.ResponseWriter, r *http.Request) {
		// https://github.com/bitmartexchange/bitmart-official-api-docs/blob/master/rest/authenticated/oauth.md#sample-response
		response := `
		{
			"access_token":"m261aeb5bfa471c67c6ac41243959ae0dd408838cdc1a47e945305dd558e2fa78",
			"expires_in":900
		}
		`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})
	m.HandleFunc("/v2/trades", func(w http.ResponseWriter, r *http.Request) {
		ts := make([]bitmart.Trade, len(trades))
		for id, t := range trades {
			// most recent first
			// note: we don't additionally subtract 1 because the trade ids
			// start at 1
			idx := len(trades) - int(id)
			if idx < 0 || idx >= len(ts) {
				reqres.RespondJSON(w, reqres.NewAPIError(
					fmt.Sprintf(
						"unexpected trades index %d for id %d (trades len %d)",
						idx, id, len(ts),
					),
					http.StatusInternalServerError,
				))
				return
			}
			ts[idx] = t
		}
		th := bitmart.TradeHistory{
			TotalTrades: int64(len(ts)),
			TotalPages:  1,
			CurrentPage: 1,
			Trades:      ts,
		}
		reqres.RespondJSON(w, reqres.OKResponse(th))
	})
	m.HandleFunc("/v2/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		idS, ok := mux.Vars(r)["id"]
		if !ok {
			reqres.RespondJSON(w, reqres.NewAPIError("id not present", http.StatusBadRequest))
			return
		}
		id, err := strconv.ParseUint(idS, 10, 64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("id must be numeric", err, http.StatusBadRequest))
			return
		}
		order, ok := orders[int64(id)]
		if !ok {
			reqres.RespondJSON(w, reqres.NewAPIError("id not found", http.StatusBadRequest))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(order))
	})

	fmt.Println("Listening for HTTP traffic on:")
	m.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pt, err := route.GetPathTemplate()
		if err == nil {
			fmt.Printf("localhost:%d%s\n", port, pt)
		}
		return nil
	})

	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), m))
}
