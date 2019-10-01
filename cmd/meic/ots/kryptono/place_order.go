package kryptono

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// order direction
const (
	SideBuy  = "BUY"
	SideSell = "SELL"
)

// ParseSide parses arbitrary text input to produce a side
func ParseSide(s string) (string, error) {
	switch strings.TrimSpace(s) {
	case SideBuy:
		return SideBuy, nil
	case SideSell:
		return SideSell, nil
	default:
		return "", fmt.Errorf("'%s' is not 'buy' or 'sell'", s)
	}
}

// PlaceOrderResponse is the response type expected when placing an order
type PlaceOrderResponse struct {
	OrderID     string `json:"order_id"`
	AccountID   string `json:"account_id"`
	OrderSymbol string `json:"order_symbol"`
	OrderSIde   string `json:"account_id"`
	Status      string `json:"status"`
	CreateTime  int64  `json:"createTime"`
	Type        string `json:"type"`
	OrderPrice  string `json:"order_price"`
	OrderSize   string `json:"order_size"`
	Executed    string `json:"executed"`
	StopPrice   string `json:"stop_price"`
	Avg         string `json:"avg"`
	Total       string `json:"total"`
}

// PlaceOrder places an order on Kryptono
func PlaceOrder(auth *Auth, symbol string, side string, price float64, amount float64) (orderID string, err error) {
	side, err = ParseSide(side)
	if err != nil {
		err = errors.Wrap(err, "invalid side")
		return
	}

	var jsdata []byte
	jsdata, err = json.Marshal(map[string]interface{}{
		"order_symbol": symbol,
		"order_side":   side,
		"order_price":  strconv.FormatFloat(price, 'f', 4, 64),
		"order_size":   strconv.FormatFloat(amount, 'f', 2, 64),
		"type":         "LIMIT",
		"timestamp":    Time(),
	})
	if err != nil {
		err = errors.Wrap(err, "json-serializing request body")
		return
	}
	fmt.Println("string jsdata = ", string(jsdata))
	buf := bytes.NewBuffer(jsdata)

	req := new(http.Request)
	ordersURL := auth.key.Endpoint + "order/add"
	req, err = http.NewRequest(http.MethodPost, ordersURL, buf)
	if err != nil {
		err = errors.Wrap(err, "constructing place order request")
		return
	}
	fmt.Println("secret = ", auth.key.Secret)
	req.Header.Set(SignatureHeader, HMACSign(auth.key.Secret, string(jsdata)))
	req.Header.Set("Content-Type", "application/json")

	resp := new(http.Response)
	resp, err = auth.Dispatch(req, 3*time.Second)
	if err != nil {
		err = errors.Wrap(err, "performing place order request")
		return
	}
	defer resp.Body.Close()

	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "reading place order response")
		return
	}
	fmt.Println("po resp = ", resp)
	fmt.Println("string data = ", string(data))

	var poresp PlaceOrderResponse
	err = json.Unmarshal(data, &poresp)
	if err != nil {
		fmt.Fprintln(os.Stderr, string(data))
		err = errors.Wrap(err, "parsing place order response")
		return
	}

	orderID = poresp.OrderID
	/* 	if poresp.Message != "" {
	   		err = errors.New(poresp.Message)
	   		err = errors.Wrap(err, "server returned error")
	   	}
	*/
	return
}
