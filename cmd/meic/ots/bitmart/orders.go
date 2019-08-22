package bitmart

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// NdauSymbol is the ordering symbol in use on bitmart
const NdauSymbol = "XND_USDT"

//go:generate stringer -type=OrderStatus

// OrderStatus is an enum defined by bitmart:
// https://github.com/bitmartexchange/bitmart-official-api-docs/blob/master/rest/authenticated/user_orders.md#status-type
type OrderStatus int64

// OrderStatus pretty names
const (
	Invalid OrderStatus = iota
	Pending
	PartialSuccess
	Success
	Canceled
	PendingAndPartialSuccess
	SuccessAndCanceled
)

var orderStatusNames map[string]OrderStatus

func init() {
	orderStatusNames = map[string]OrderStatus{
		"pending":                  Pending,
		"partialsuccess":           PartialSuccess,
		"success":                  Success,
		"canceled":                 Canceled,
		"pendingandpartialsuccess": PendingAndPartialSuccess,
		"successandcanceled":       SuccessAndCanceled,
	}
}

// OrderStatusFrom gets an order status from its name
func OrderStatusFrom(s string) OrderStatus {
	var out OrderStatus // invalid by default
	if os, ok := orderStatusNames[strings.ToLower(s)]; ok {
		out = os
	}
	return out
}

// Order represents a particular order
type Order struct {
	EntrustID       int64   `json:"entrust_id"`
	Symbol          string  `json:"symbol"`
	Timestamp       int64   `json:"timestamp"` // milliseconds since unix epoch
	Side            string  `json:"side"`
	Price           float64 `json:"price"`
	Fees            float64 `json:"fees"`
	OriginalAmount  float64 `json:"original_amount"`
	ExecutedAmount  float64 `json:"executed_amount"`
	RemainingAmount float64 `json:"remaining_amount"`
	Status          int64   `json:"status"`
}

// IsSale is true when the order is a sell order
func (t *Order) IsSale() bool {
	s, err := ParseSide(t.Side)
	if err != nil {
		return false
	}
	return s == SideSell
}

// IsBuy is true when the order is a buy order
func (t *Order) IsBuy() bool {
	s, err := ParseSide(t.Side)
	if err != nil {
		return false
	}
	return s == SideBuy
}

// UnmarshalJSON implements json.Unmarshaler
//
// It's necessary because we can't trust bitmart to actually send numbers
// in numeric fields
func (t *Order) UnmarshalJSON(data []byte) error {
	var obj map[string]interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return errors.Wrap(err, "wallet")
	}

	// attempts to get all fields. err remains nil if everything succeeded
	t.EntrustID, err = getInt(obj, "entrust_id")
	t.Symbol, err = getStr(obj, "symbol")
	t.Timestamp, err = getInt(obj, "timestamp")
	t.Side, err = getStr(obj, "side")
	t.Price, err = getFloat(obj, "price")
	t.Fees, err = getFloat(obj, "fees")
	t.OriginalAmount, err = getFloat(obj, "original_amount")
	t.ExecutedAmount, err = getFloat(obj, "executed_amount")
	t.RemainingAmount, err = getFloat(obj, "remaining_amount")
	t.Status, err = getInt(obj, "status")
	return err
}

// OrderHistory is the response from the bitmart order history request
type OrderHistory struct {
	Orders      []Order `json:"orders"`
	TotalPages  int64   `json:"total_pages"`
	TotalOrders int64   `json:"total_orders"`
	CurrentPage int64   `json:"current_page"`
}

// GetOrderHistory retrieves the list of all user orders
func GetOrderHistory(auth *Auth, symbol string, status OrderStatus) ([]Order, error) {
	if status == Invalid {
		return nil, errors.New("invalid status")
	}
	if symbol == "" {
		return nil, errors.New("symbol must not be empty")
	}

	var offset = 0
	const limit = 1000
	var th OrderHistory
	orders := make([]Order, 0, limit)

	getPage := func() error {
		queryParams := url.Values{}
		queryParams.Set("symbol", url.QueryEscape(symbol))
		queryParams.Set("status", fmt.Sprintf("%d", status))
		queryParams.Set("offset", fmt.Sprintf("%d", offset))
		queryParams.Set("limit", fmt.Sprintf("%d", limit))

		ordersURL := auth.key.Endpoint + "orders"
		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s?%s", ordersURL, queryParams.Encode()),
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "constructing order history request")
		}

		resp, err := auth.Dispatch(req, 5*time.Second)
		if err != nil {
			return errors.Wrap(err, "performing order history request")
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "reading order history response")
		}

		err = json.Unmarshal(data, &th)
		if err != nil {
			return errors.Wrap(err, "parsing order history response")
		}
		return nil
	}

	// get first page
	err := getPage()
	if err != nil {
		return nil, errors.Wrap(err, "getting first order history page")
	}
	orders = append(orders, th.Orders...)

	// in the future, parallelize this?
	for th.CurrentPage < th.TotalPages {
		offset += limit
		err = getPage()
		if err != nil {
			return orders, errors.Wrap(err, fmt.Sprintf("getting order history page %d", offset/limit))
		}
		orders = append(orders, th.Orders...)
	}

	return orders, nil
}

// GetOrder gets the named order
func GetOrder(auth *Auth, entrustID int64) (*Order, error) {
	ordersURL := auth.key.Endpoint + "orders"
	url := fmt.Sprintf("%s/%d", ordersURL, entrustID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "constructing order request")
	}
	resp, err := auth.Dispatch(req, 2*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "performing order request")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading order response")
	}

	var order Order
	err = json.Unmarshal(data, &order)
	if err != nil {
		return nil, errors.Wrap(err, "parsing order response")
	}

	return &order, nil
}
