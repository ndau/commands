package bitmart

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

// Some headers and header-adjacent items for bitmart
const (
	AuthHeader      = "x-bm-authorization"
	TimestampHeader = "x-bm-timestamp"
	BearerPrefix    = "Bearer "
)

// Time returns the number of milliseconds since unix epoch (utc)
func Time() int64 {
	return time.Now().UnixNano() / 1000000
	//               nano to micro  ^^^
	//              micro to milli     ^^^
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
	err := a.refreshToken()
	if err != nil {
		return errors.Wrap(err, "authenticating")
	}

	// if the API key specifies a replacement endpoint, use it
	a.key.SubsURL(request.URL)

	request.Header.Set(AuthHeader, BearerPrefix+a.token.Access)
	request.Header.Set(TimestampHeader, fmt.Sprintf("%d", Time()))
	return nil
}

// Dispatch an authorized request after adding appropriate authentication headers.
func (a *Auth) Dispatch(request *http.Request, timeout time.Duration) (resp *http.Response, err error) {
	err = a.Authorize(request)
	if err != nil {
		err = errors.Wrap(err, "authorizing dispatch")
		return
	}
	a.client.Timeout = timeout
	fmt.Fprintf(
		os.Stderr,
		"debug: sending %s request to %s\n",
		request.Method, request.URL,
	)
	spew.Fdump(os.Stderr, "debug: headers:", request.Header)
	return a.client.Do(request)
}
