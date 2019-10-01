package kryptono

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

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
	HexID        string  `json:"hex_id"`
	OrderID      string  `json:"order_id"`
	AccountID    string  `json:"account_id"`
	OrderSymbol  string  `json:"order_symbol"`
	OrderSide    string  `json:"order_side"`
	Status       string  `json:"status"`
	CreateTime   int64   `json:"createTime"`
	Type         string  `json:"type"`
	TimeMatching int64   `json:"timeMatching"`
	OrderPrice   float64 `json:"order_price"`
	OrderSize    float64 `json:"order_size"`
	Executed     float64 `json:"executed"`
	StopPrice    float64 `json:"stop_price"`
	Avg          float64 `json:"avg"`
	Total        string  `json:"total"`
}

// IsSale is true when the order is a sell order
func (t *Order) IsSale() bool {
	s, err := ParseSide(t.OrderSide)
	if err != nil {
		return false
	}
	return s == SideSell
}

// IsBuy is true when the order is a buy order
func (t *Order) IsBuy() bool {
	s, err := ParseSide(t.OrderSide)
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
	t.HexID, err = getStr(obj, "hex_id")
	t.OrderID, err = getStr(obj, "order_id")
	t.AccountID, err = getStr(obj, "account_id")
	t.OrderSymbol, err = getStr(obj, "order_symbol")
	t.OrderSide, err = getStr(obj, "order_side")
	t.Status, err = getStr(obj, "status")
	t.CreateTime, err = getInt(obj, "createTime")
	t.Type, err = getStr(obj, "type")
	t.TimeMatching, err = getInt(obj, "timeMatching")
	t.OrderPrice, err = getFloat(obj, "order_price")
	t.OrderSize, err = getFloat(obj, "order_size")
	t.Executed, err = getFloat(obj, "executed")
	t.StopPrice, err = getFloat(obj, "stop_price")
	t.Avg, err = getFloat(obj, "avg")
	t.Total, err = getStr(obj, "total")

	return err
}

// GetOrderHistory retrieves the list of all user orders
func GetOrderHistory(auth *Auth, symbol string, status OrderStatus) ([]Order, error) {
	if status == Invalid {
		return nil, errors.New("invalid status")
	}
	if symbol == "" {
		return nil, errors.New("symbol must not be empty")
	}

	// var offset = 0
	const limit = 1000
	var th []Order
	orders := make([]Order, 0, limit)

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

		fmt.Println("buf = ", buf)

		req := new(http.Request)
		ordersURL := auth.key.Endpoint + "order/list/all"
		req, err = http.NewRequest(http.MethodPost, ordersURL, buf)
		if err != nil {
			err = errors.Wrap(err, "constructing order request")
			return err
		}
		req.Header.Set(SignatureHeader, HMACSign(auth.key.Secret, string(jsdata)))
		req.Header.Set("Content-Type", "application/json")

		fmt.Println("string jsdata = ", string(jsdata))

		resp := new(http.Response)
		resp, err = auth.Dispatch(req, 3*time.Second)
		if err != nil {
			err = errors.Wrap(err, "performing order request")
			return err
		}
		defer resp.Body.Close()
		fmt.Println("resp = ", resp)

		var data []byte
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = errors.Wrap(err, "reading order response")
			return err
		}

		fmt.Println("data = ", data)

		fmt.Println("string data = ", string(data))

		err = json.Unmarshal(data, &th)
		fmt.Println("th = ", th)
		if err != nil {
			fmt.Fprintln(os.Stderr, string(data))
			err = errors.Wrap(err, "parsing order response")
			return err
		}

		return nil
	}

	// get first page
	err := getPage()
	if err != nil {
		return nil, errors.Wrap(err, "getting first order history page")
	}
	orders = append(orders, th...)

	// in the future, parallelize this?
	/* 	for th.CurrentPage < th.TotalPages {
	   		offset += limit
	   		err = getPage()
	   		if err != nil {
	   			return orders, errors.Wrap(err, fmt.Sprintf("getting order history page %d", offset/limit))
	   		}
	   		orders = append(orders, th.Orders...)
	   	}
	*/
	return orders, nil
}

// GetOrder gets the named order
func GetOrder(auth *Auth, hexID string) (*Order, error) {
	ordersURL := auth.key.Endpoint + "orders"
	url := fmt.Sprintf("%s/%s", ordersURL, hexID)
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
