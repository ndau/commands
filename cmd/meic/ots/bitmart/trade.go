package bitmart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Trade represents a particular trade
type Trade struct {
	Symbol    string    `json:"symbol"`
	Amount    math.Ndau `json:"amount"`
	Fees      float64   `json:"fees"`
	TradeID   int64     `json:"trade_id"`
	Price     float64   `json:"price"`
	Active    bool      `json:"active"`
	EntrustID int64     `json:"entrust_id"`
	Timestamp int64     `json:"timestamp"` // milliseconds since unix epoch
}

// UnmarshalJSON implements json.Unmarshaler
//
// It's necessary because we can't trust bitmart to actually send numbers
// in numeric fields
func (t *Trade) UnmarshalJSON(data []byte) error {
	var obj map[string]interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return errors.Wrap(err, "wallet")
	}

	// attempts to get all fields. err remains nil if everything succeeded
	t.Symbol, err = getStr(obj, "symbol")
	t.Amount, err = getNdau(obj, "amount")
	t.Fees, err = getFloat(obj, "fees")
	t.TradeID, err = getInt(obj, "trade_id")
	t.Price, err = getFloat(obj, "price")
	t.Active, err = getBool(obj, "active")
	t.EntrustID, err = getInt(obj, "entrust_id")
	t.Timestamp, err = getInt(obj, "timestamp")
	return err
}

// MarshalJSON implements json.Marshaler
//
// It's necessary so that when mocking out the bitmart service, we don't end up
// with "amount" values which are just way too high
func (t *Trade) MarshalJSON() ([]byte, error) {
	// make an alias to avoid recursion
	type Alias Trade
	return json.Marshal(&struct {
		Amount string `json:"amount"`
		*Alias
	}{
		Amount: t.Amount.String(),
		Alias:  (*Alias)(t),
	})
}

// TradeHistory is the response from the bitmart trade history request
type TradeHistory struct {
	TotalTrades int64   `json:"total_trades"`
	TotalPages  int64   `json:"total_pages"`
	CurrentPage int64   `json:"current_page"`
	Trades      []Trade `json:"trades"`
}

// GetTradeHistory retrieves the list of all user trades
func GetTradeHistory(auth *Auth, symbol string) ([]Trade, error) {
	return GetTradeHistoryAfter(auth, symbol, 0)
}

// GetTradeHistoryAfter retrieves the list of all user trades whose trade_id is
// greater than tradeIDLimit.
func GetTradeHistoryAfter(auth *Auth, symbol string, tradeIDLimit int64) ([]Trade, error) {
	if symbol == "" {
		return nil, errors.New("symbol must not be empty")
	}
	var offset = 0
	const limit = 1000
	var th TradeHistory
	trades := make([]Trade, 0, limit)
	stop := false

	getPage := func() error {
		queryParams := url.Values{}
		queryParams.Set("symbol", url.QueryEscape(symbol))
		queryParams.Set("offset", fmt.Sprintf("%d", offset))
		queryParams.Set("limit", fmt.Sprintf("%d", limit))

		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s?%s", APITrades, queryParams.Encode()),
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "constructing trade history request")
		}

		resp, err := auth.Dispatch(req, 5*time.Second)
		if err != nil {
			return errors.Wrap(err, "performing trade history request")
		}
		defer resp.Body.Close()

		// log.Print(resp)

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "reading trade history response")
		}

		var out bytes.Buffer
		err = json.Indent(&out, data, "", "  ")
		log.Printf("data = %s", out.Bytes())

		offset += limit

		err = json.Unmarshal(data, &th)
		if err != nil {
			return errors.Wrap(err, "parsing trade history response")
		}

		midx := -1
		for idx, trade := range th.Trades {
			if trade.TradeID > tradeIDLimit {
				midx = idx
			} else {
				stop = true
				break
			}
		}
		trades = append(trades, th.Trades[:midx+1]...)
		log.Print(trades)
		return nil
	}

	// get first page
	err := getPage()
	if err != nil {
		return nil, errors.Wrap(err, "getting first trade history page")
	}

	for !stop && th.CurrentPage < th.TotalPages {
		err = getPage()
		if err != nil {
			return trades, errors.Wrap(err, fmt.Sprintf("getting trade history page %d", offset/limit))
		}
	}

	return trades, nil
}

// FilterSales retains only those Trades which correspond to a sell order
//
// For each trade, dispatches an individual goroutine fetching the appropriate
// Order. If that Order comes back as a sell order, retains the Trade; otherwise,
// discards it.
//
// This is a fairly IO-intensive call, though it's made as concurrent as possible.
func FilterSales(auth *Auth, trades []Trade) ([]Trade, error) {
	// order traide pair
	type otp struct {
		order Order
		trade Trade
	}
	orders := make(chan otp)
	errs := make(chan error)
	for _, trade := range trades {
		go func(trade Trade) {
			order, err := GetOrder(auth, trade.EntrustID)
			// always send exactly one message on a channel
			if err == nil {
				if order != nil {
					orders <- otp{*order, trade}
				} else {
					errs <- errors.New("nil order returned from GetOrder")
				}
			} else {
				errs <- errors.Wrap(err, fmt.Sprint(trade.EntrustID))
			}
		}(trade)
	}

	filtered := make([]Trade, 0, len(trades))

	allerrs := make([]string, 0)

	deadline := time.After(5 * time.Second)
	for responses := 0; responses < len(trades); responses++ {
		select {
		case pair := <-orders:
			if pair.order.IsSale() {
				filtered = append(filtered, pair.trade)
			}
		case err := <-errs:
			allerrs = append(allerrs, err.Error())
		case <-deadline:
			allerrs = append(allerrs, "timeout: deadline expired")
			break
		}
	}

	var err error
	if len(allerrs) > 0 {
		err = errors.New(strings.Join(allerrs, "; "))
	}
	return filtered, err
}
