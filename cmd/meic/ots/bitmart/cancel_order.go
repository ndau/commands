package bitmart

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// CancelOrder cancels an order on bitmart
func CancelOrder(auth *Auth, entrustID uint64) (err error) {

	err = nil
	url := auth.key.Endpoint + "orders/" + strconv.FormatUint(entrustID, 10)
	fmt.Println("url = ", url)
	req := new(http.Request)
	req, err = http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		err = errors.Wrap(err, "constructing order request")
		return
	}

	message := fmt.Sprintf("entrust_id=%s", strconv.FormatUint(entrustID, 10))

	req.Header.Set(SignatureHeader, HMACSign(auth.key.Secret, message))
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("req = ", req)
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
	fmt.Println(string(data))

	return
}
