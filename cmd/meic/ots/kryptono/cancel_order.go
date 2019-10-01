package kryptono

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// CancelOrder cancels an order on bitmart
func CancelOrder(auth *Auth, orderID string, symbol string) (err error) {
	var jsdata []byte
	jsdata, err = json.Marshal(map[string]interface{}{
		"order_id":     orderID,
		"order_symbol": symbol,
		"timestamp":    Time(),
	})
	if err != nil {
		err = errors.Wrap(err, "json-serializing request body")
		return err
	}
	fmt.Println("jsdata = ", jsdata)
	buf := bytes.NewBuffer(jsdata)

	url := auth.key.Endpoint + "order/cancel"
	fmt.Println("url = ", url)
	req := new(http.Request)
	req, err = http.NewRequest(http.MethodDelete, url, buf)
	if err != nil {
		err = errors.Wrap(err, "constructing order request")
		return
	}

	req.Header.Set(SignatureHeader, HMACSign(auth.key.Secret, string(jsdata)))
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("string jsdata = ", string(jsdata))

	fmt.Println("req = ", req)
	resp := new(http.Response)
	resp, err = auth.Dispatch(req, 3*time.Second)
	if err != nil {
		err = errors.Wrap(err, "performing cancel request")
		return
	}
	defer resp.Body.Close()

	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "reading cancel response")
		return
	}
	fmt.Println("cancel resp = ", string(data))

	return
}
