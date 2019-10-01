package kryptono

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
)

// Some headers and header-adjacent items for bitmart
const (
	AuthHeader = "Authorization"
)

// TimeResponse is the response type expected when placing an order
type TimeResponse struct {
	ServerTime int64 `json:"server_time"`
}

// Time returns the number of milliseconds since unix epoch (utc)
func Time() int64 {
	// return time.Now().UnixNano() / 1000000
	//               nano to micro  ^^^
	//              micro to milli     ^^^
	req, err := http.NewRequest(http.MethodGet, "https://testenv1.kryptono.exchange/k/api/v2/time", nil)
	if err != nil {
		err = errors.Wrap(err, "creating req")
		return 0
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = errors.Wrap(err, "authorizing dispatch")
		return 0
	}
	fmt.Println("resp = ", resp)
	defer resp.Body.Close()

	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "reading order response")
		return 0
	}

	var tiresp TimeResponse
	err = json.Unmarshal(data, &tiresp)
	if err != nil {
		fmt.Fprintln(os.Stderr, string(data))
		err = errors.Wrap(err, "parsing order response")
		return 0
	}

	fmt.Println("time = ", tiresp)
	return int64(tiresp.ServerTime)
}

// Auth is a helper which persists and requests new auth tokens, transparently
type Auth struct {
	key    APIKey
	token  *Token
	client *http.Client
}

// NewAuth creates a new Auth
func NewAuth(key APIKey) Auth {
	return Auth{
		key:    key,
		client: http.DefaultClient,
	}
}

// ensure that we have a valid auth token, idempotently
func (a *Auth) refreshToken() (err error) {
	// if we're within 5 seconds of expiry, just refresh anyway, to be safe
	if a.token != nil && a.token.expiry.After(time.Now().Add(5*time.Second)) {
		return
	}
	a.token, err = a.key.Authenticate()
	return
}

// Authorize modifies a request by adding appropriate authentication headers.
//
// Note that Bitmart's API includes time-sensitive headers; requests should
// be dispatched without delay.
func (a *Auth) Authorize(request *http.Request) error {
	// err := a.refreshToken()
	// if err != nil {
	// 	return errors.Wrap(err, "authenticating")
	// }

	// if the API key specifies a replacement endpoint, use it
	request.Host = a.key.SubsURL(request.URL)

	request.Header.Set(AuthHeader, a.key.Access)
	return nil
}

// Dispatch an authorized request after adding appropriate authentication headers.
func (a *Auth) Dispatch(request *http.Request, timeout time.Duration) (resp *http.Response, err error) {
	err = a.Authorize(request)
	if err != nil {
		err = errors.Wrap(err, "authorizing dispatch")
		return
	}
	fmt.Println("request = ", request)
	a.client.Timeout = timeout
	return a.client.Do(request)
}
