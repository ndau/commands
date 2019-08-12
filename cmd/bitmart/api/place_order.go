package bitmart

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
	SideBuy  = "buy"
	SideSell = "sell"
)

// SignatureHeader names the key used for signing the request body per the Bitmart strategy
const SignatureHeader = "x-bm-signature"

// ParseSide parses arbitrary text input to produce a side
func ParseSide(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
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
	EntrustID int64  `json:"entrust_id"`
	Message   string `json:"message,omitempty"`
}

// bitmart uses a very particular encoding scheme to validate request bodies
// https://github.com/bitmartexchange/bitmart-official-api-docs/blob/master/rest/authenticated/post_order.md
//
// returns just the signature, not the full header
func prepareOrderSignature(auth *Auth, symbol string, side string, price float64, amount float64) string {
	msg := fmt.Sprintf(
		"amount=%s&price=%s&side=%s&symbol=%s",
		strconv.FormatFloat(amount, 'f', -1, 64),
		strconv.FormatFloat(price, 'f', -1, 64),
		side,
		symbol,
	)
	return HMACSign(auth.key.Secret, msg)
}

// PlaceOrder places an order on bitmart
func PlaceOrder(auth *Auth, symbol string, side string, price float64, amount float64) (entrustID int64, err error) {
	side, err = ParseSide(side)
	if err != nil {
		err = errors.Wrap(err, "invalid side")
		return
	}

	var jsdata []byte
	jsdata, err = json.Marshal(map[string]interface{}{
		"amount": amount,
		"price":  price,
		"side":   side,
		"symbol": symbol,
	})
	if err != nil {
		err = errors.Wrap(err, "json-serializing request body")
		return
	}
	buf := bytes.NewBuffer(jsdata)

	req := new(http.Request)
	req, err = http.NewRequest(http.MethodPost, APIOrders, buf)
	if err != nil {
		err = errors.Wrap(err, "constructing order request")
		return
	}
	req.Header.Set(SignatureHeader, prepareOrderSignature(auth, symbol, side, price, amount))
	req.Header.Set("Content-Type", "application/json")

	resp := new(http.Response)
	resp, err = auth.Dispatch(req, 3*time.Second)
	if err != nil {
		err = errors.Wrap(err, "performing order request")
		return
	}
	defer resp.Body.Close()

	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "reading order response")
		return
	}

	var poresp PlaceOrderResponse
	err = json.Unmarshal(data, &poresp)
	if err != nil {
		fmt.Fprintln(os.Stderr, string(data))
		err = errors.Wrap(err, "parsing order response")
		return
	}

	entrustID = poresp.EntrustID
	if poresp.Message != "" {
		err = errors.New(poresp.Message)
		err = errors.Wrap(err, "server returned error")
	}

	return
}
