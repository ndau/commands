package kryptono

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Trade represents a particular trade
type Trade struct {
	HexID     string    `json:"hex_id"`
	Symbol    string    `json:"symbol"`
	OrderID   string    `json:"order_id"`
	OrderSide string    `json:"order_side"`
	Price     float64   `json:"price"`
	Quantity  math.Ndau `json:"quantity"`
	Fee       string    `json:"fee"`
	Total     string    `json:"total"`
	Timestamp int64     `json:"timestamp"`
}

// UnmarshalJSON implements json.Unmarshaler
//
// It's necessary because we can't trust kryptono to actually send numbers
// in numeric fields
func (t *Trade) UnmarshalJSON(data []byte) error {
	var obj map[string]interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return errors.Wrap(err, "wallet")
	}

	// attempts to get all fields. err remains nil if everything succeeded
	t.HexID, err = getStr(obj, "hex_id")
	t.Symbol, err = getStr(obj, "symbol")
	t.OrderID, err = getStr(obj, "order_id")
	t.OrderSide, err = getStr(obj, "order_side")
	t.Price, err = getFloat(obj, "price")
	t.Quantity, err = getNdau(obj, "quantity")
	t.Fee, err = getStr(obj, "fee")
	t.Total, err = getStr(obj, "total")
	t.Timestamp, err = getInt(obj, "timestamp")

	return err
}

// MarshalJSON implements json.Marshaler
//
// It's necessary so that when mocking out the kryptono service, we don't end up
// with "amount" values which are just way too high
/* func (t *Trade) MarshalJSON() ([]byte, error) {
	// make an alias to avoid recursion
	type Alias Trade
	return json.Marshal(&struct {
		Amount string `json:"amount"`
		*Alias
	}{
		Amount: t.Quantity.String(),
		Alias:  (*Alias)(t),
	})
}
*/
// TradeHistory is the response from the bitmart trade history request
type TradeHistory struct {
	TotalTrades int64   `json:"total_trades"`
	TotalPages  int64   `json:"total_pages"`
	CurrentPage int64   `json:"current_page"`
	Trades      []Trade `json:"trades"`
}

// GetTradeHistory retrieves the list of all user trades
func GetTradeHistory(auth *Auth, symbol string) ([]Trade, string, error) {
	return GetTradeHistoryAfter(auth, symbol, "")
}

// GetTradeHistoryAfter retrieves the list of all user trades whose trade_id is
// greater than tradeIDLimit.
func GetTradeHistoryAfter(auth *Auth, symbol string, tradeIDLimit string) ([]Trade, string, error) {
	if symbol == "" {
		return nil, "", errors.New("symbol must not be empty")
	}
	var offset = 0
	const limit = 1000
	var th []Trade
	trades := make([]Trade, 0, limit)
	stop := false
	maxTradeID := tradeIDLimit

	getPage := func() error {
		var jsdata []byte
		var err error
		jsdata, err = json.Marshal(map[string]interface{}{
			"symbol":    symbol,
			"timestamp": Time(),
		})
		if err != nil {
			err = errors.Wrap(err, "json-serializing request body")
			return err
		}
		fmt.Println("jsdata = ", jsdata)
		buf := bytes.NewBuffer(jsdata)

		req := new(http.Request)
		tradesURL := auth.key.Endpoint + "order/list/trades"
		req, err = http.NewRequest(http.MethodPost, tradesURL, buf)
		if err != nil {
			err = errors.Wrap(err, "constructing trades request")
			return err
		}

		req.Header.Set(SignatureHeader, HMACSign(auth.key.Secret, string(jsdata)))
		req.Header.Set("Content-Type", "application/json")

		fmt.Println("string jsdata = ", string(jsdata))

		resp, err := auth.Dispatch(req, 3*time.Second)
		if err != nil {
			return errors.Wrap(err, "performing trade history request")
		}
		defer resp.Body.Close()

		// log.Print(resp)

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "reading trade history response")
		}

		fmt.Println("string data = ", string(data))

		var out bytes.Buffer
		err = json.Indent(&out, data, "", "  ")
		//		log.Printf("data = %s", out.Bytes())

		offset += limit

		err = json.Unmarshal(data, &th)
		if err != nil {
			return errors.Wrap(err, "parsing trade history response")
		}

		//		log.Println("trade hist = ", th)
		midx := -1
		for idx, trade := range th {
			if trade.HexID > tradeIDLimit {
				midx = idx
				if trade.HexID > maxTradeID {
					maxTradeID = trade.HexID
				}
			} else {
				stop = true
				break
			}
		}
		trades = append(trades, th[:midx+1]...)
		if len(trades) > 0 {
			log.Println("trades = ", trades)
		}
		return nil
	}

	// get first page
	err := getPage()
	if err != nil {
		return nil, maxTradeID, errors.Wrap(err, "getting first trade history page")
	}

	/* 	for !stop && th.CurrentPage < (th.TotalPages-1) {
	   		err = getPage()
	   		if err != nil {
	   			return trades, maxTradeID, errors.Wrap(err, fmt.Sprintf("getting trade history page %d", offset/limit))
	   		}
	   	}
	*/
	return trades, maxTradeID, nil
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
			order, err := GetOrder(auth, trade.HexID)
			// always send exactly one message on a channel
			if err == nil {
				if order != nil {
					orders <- otp{*order, trade}
				} else {
					errs <- errors.New("nil order returned from GetOrder")
				}
			} else {
				errs <- errors.Wrap(err, fmt.Sprint(trade.HexID))
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
